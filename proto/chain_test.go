package proto

import (
	"testing"
)

func TestChain(t *testing.T) {
	sm1 := NewSimpleMatcher(HTTPMethods, nil)
	sm2 := NewSimpleMatcher([][]byte{[]byte("GET")}, nil)
	sm3 := NewSimpleMatcher([][]byte{[]byte("GAT")}, nil)

	cm1 := NewChain(nil, sm1.Check, sm2.Check)
	cm2 := NewChain(nil, sm1.Check, sm3.Check)

	match, _ := cm1.Check([]byte("GET"), nil)
	if !match {
		t.Errorf("Chain did not match when it should")
	}

	match, _ = cm1.Check([]byte("GAT"), nil)
	if match {
		t.Errorf("Chain matched when it shouldn't")
	}

	match, _ = cm2.Check([]byte("GET"), nil)
	if match {
		t.Errorf("Chain matched when it shouldn't")
	}

}
