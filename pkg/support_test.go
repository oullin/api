package pkg

import (
	"errors"
	"testing"
)

type fakeCloser struct {
	closed bool
	err    error
}

func (f *fakeCloser) Close() error {
	f.closed = true
	return f.err
}

func TestCloseWithLog(t *testing.T) {
	c := &fakeCloser{}
	CloseWithLog(c)
	if !c.closed {
		t.Fatalf("close not called")
	}

	c2 := &fakeCloser{err: errors.New("fail")}
	CloseWithLog(c2)
	if !c2.closed {
		t.Fatalf("close not called with error")
	}
}
