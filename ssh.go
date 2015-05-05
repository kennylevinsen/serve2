package serve2

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"

	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"bytes"
	"errors"
	"github.com/kr/pty"
)

// SSHProtoHandler allows for secure shell login for the user of the server
type SSHProtoHandler struct {
	hostPrivateKeySigner ssh.Signer
	sshConfig            ssh.ServerConfig
	allowedKeys          [][]byte
	defaultShell         string
}

// Setup configures the SSHProtoHandlers private server key, as well as authorized_keys to permit login from
func (s *SSHProtoHandler) Setup(hostKey, authKeys string) {

	hostPrivateKey, err := ioutil.ReadFile(hostKey)
	if err != nil {
		panic(err)
	}

	s.hostPrivateKeySigner, err = ssh.ParsePrivateKey(hostPrivateKey)
	if err != nil {
		panic(err)
	}

	s.sshConfig = ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			k := key.Marshal()
			for _, pk := range s.allowedKeys {
				if bytes.Compare(k, pk) == 0 {
					// OK!
					return nil, nil
				}
			}
			return nil, errors.New("unknown publickey")
		},
	}
	s.sshConfig.AddHostKey(s.hostPrivateKeySigner)

	authFile, err := ioutil.ReadFile(authKeys)
	if err != nil {
		panic(err)
	}

	for len(authFile) > 0 {
		var pk ssh.PublicKey
		pk, _, _, authFile, err = ssh.ParseAuthorizedKey(authFile)
		if err != nil {
			panic(err)
		}
		s.allowedKeys = append(s.allowedKeys, pk.Marshal())
	}

}

// Handle creats a SSH session from the connection
func (s *SSHProtoHandler) Handle(c net.Conn) net.Conn {
	sshConn, chans, reqs, err := ssh.NewServerConn(c, &s.sshConfig)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil
	}
	log.Printf("serving SSH (%s, %s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	go s.handleRequests(reqs)
	go s.handleChannels(chans)

	return nil
}

func (s *SSHProtoHandler) Check(header []byte) bool {
	return string(header[:7]) == "SSH-2.0"
}

func (s *SSHProtoHandler) BytesRequired() int {
	return 7
}

func (s *SSHProtoHandler) handleRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		log.Printf("recieved out-of-band request: %+v", req)
	}
}

// NewSSHProtoHandler returns a fully initialized SSHProtoHandler
func NewSSHProtoHandler(hostKey, authKeys string) *SSHProtoHandler {
	s := SSHProtoHandler{
		defaultShell: "zsh",
	}

	s.Setup(hostKey, authKeys)
	return &s
}

//
// XXX: Beware, beneath lies dragons and copy-pasta
//

// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
func ptyRun(c *exec.Cmd, tty *os.File) (err error) {
	defer tty.Close()
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	c.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
	return c.Start()
}

func (s *SSHProtoHandler) handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("could not accept channel (%s)", err)
			continue
		}

		// allocate a terminal for this channel
		log.Print("creating pty...")
		// Create new pty
		f, tty, err := pty.Open()
		if err != nil {
			log.Printf("could not start pty (%s)", err)
			continue
		}

		// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "exec":
					log.Println("exec request")
					ok = true
					command := string(req.Payload[4 : req.Payload[3]+4])
					cmd := exec.Command(s.defaultShell, []string{"-c", command}...)

					cmd.Stdout = channel
					cmd.Stderr = channel
					cmd.Stdin = channel

					err := cmd.Start()
					if err != nil {
						log.Printf("could not start command (%s)", err)
						continue
					}

					// teardown session
					go func() {
						_, err := cmd.Process.Wait()
						if err != nil {
							log.Printf("failed to exit bash (%s)", err)
						}
						channel.Close()
						log.Printf("session closed")
					}()
				case "shell":
					log.Println("shell request")
					cmd := exec.Command(s.defaultShell)
					cmd.Env = []string{"TERM=xterm"}
					err := ptyRun(cmd, tty)
					if err != nil {
						log.Printf("%s", err)
					}

					// Teardown session
					var once sync.Once
					close := func() {
						channel.Close()
						log.Printf("session closed")
					}

					// Pipe session to bash and visa-versa
					go func() {
						io.Copy(channel, f)
						once.Do(close)
					}()

					go func() {
						io.Copy(f, channel)
						once.Do(close)
					}()

					// We don't accept any commands (Payload),
					// only the default shell.
					if len(req.Payload) == 0 {
						ok = true
					}
				case "pty-req":
					log.Println("pty request")
					// Responding 'ok' here will let the client
					// know we have a pty ready for input
					ok = true
					// Parse body...
					termLen := req.Payload[3]
					termEnv := string(req.Payload[4 : termLen+4])
					w, h := parseDims(req.Payload[termLen+4:])
					setWinsize(f.Fd(), w, h)
					log.Printf("pty-req '%s'", termEnv)
				case "window-change":
					log.Println("window change request")
					w, h := parseDims(req.Payload)
					setWinsize(f.Fd(), w, h)
					continue //no response
				}

				if !ok {
					log.Printf("declining %s request...", req.Type)
				}

				req.Reply(ok, nil)
			}
		}(requests)
	}
}

// parseDims extracts two uint32s from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// winsize stores the Height and Width of a terminal.
type winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// setWinsize sets the size of the given pty.
func setWinsize(fd uintptr, w, h uint32) {
	ws := &winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
