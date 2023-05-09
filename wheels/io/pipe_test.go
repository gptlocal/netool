package io_test

import (
	. "github.com/gptlocal/netool/wheels/io"
	"io"
	"testing"
)

type pipeReturn struct {
	n   int
	err error
}

func TestWriteEmpty(t *testing.T) {
	r, w := Pipe()
	go func() {
		w.Write([]byte{})
		w.Close()
	}()
	var b [2]byte
	n, err := io.ReadFull(r, b[0:2])
	t.Logf("read %d err: %v", n, err)
	r.Close()
}

func TestWriteNil(t *testing.T) {
	r, w := Pipe()
	go func() {
		w.Write(nil)
		w.Close()
	}()
	var b [2]byte
	n, err := io.ReadFull(r, b[0:2])
	t.Logf("read %d err: %v", n, err)
	r.Close()
}

func TestWriteAfterWriterClose(t *testing.T) {
	r, w := Pipe()

	done := make(chan bool)
	var writeErr error
	go func() {
		_, err := w.Write([]byte("hello"))
		if err != nil {
			t.Errorf("got error: %q; expected none", err)
		}
		w.Close()
		_, writeErr = w.Write([]byte("world"))
		done <- true
	}()

	buf := make([]byte, 100)
	var result string
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		t.Fatalf("got: %q; want: %q", err, io.ErrUnexpectedEOF)
	}
	result = string(buf[0:n])
	t.Logf("read %d, result: %s, err: %v", n, result, err)

	<-done

	if result != "hello" {
		t.Errorf("got: %q; want: %q", result, "hello")
	}
	if writeErr != ErrClosedPipe {
		t.Errorf("got: %q; want: %q", writeErr, ErrClosedPipe)
	}
}

func TestPipeCloseError(t *testing.T) {
	type testError1 struct{ error }
	type testError2 struct{ error }

	r, w := Pipe()
	r.(*PipeReader).CloseWithError(testError1{})
	if _, err := w.Write(nil); err != (testError1{}) {
		t.Errorf("Write error: got %T, want testError1", err)
	}
	r.(*PipeReader).CloseWithError(testError2{})
	if _, err := w.Write(nil); err != (testError1{}) {
		t.Errorf("Write error: got %T, want testError1", err)
	}

	r, w = Pipe()
	w.(*PipeWriter).CloseWithError(testError1{})
	if _, err := r.Read(nil); err != (testError1{}) {
		t.Errorf("Read error: got %T, want testError1", err)
	}
	w.(*PipeWriter).CloseWithError(testError2{})
	if _, err := r.Read(nil); err != (testError1{}) {
		t.Errorf("Read error: got %T, want testError1", err)
	}
}

func TestPipe1(t *testing.T) {
	c := make(chan int)
	r, w := Pipe()
	defer r.Close()
	defer w.Close()

	go checkWrite(t, w, []byte("hello, world"), c)

	var buf = make([]byte, 64)
	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("read: %v", err)
	} else if n != 12 || string(buf[0:12]) != "hello, world" {
		t.Errorf("bad read: got %q", buf[0:n])
	}
	t.Logf("read %d bytes: %q", n, buf[0:n])
	<-c
}

func TestPipe2(t *testing.T) {
	c := make(chan int)
	r, w := Pipe()
	go reader(t, r, c)

	var buf = make([]byte, 64)
	for i := 0; i < 5; i++ {
		p := buf[0 : 5+i*10]
		n, err := w.Write(p)
		if n != len(p) {
			t.Errorf("wrote %d, got %d", len(p), n)
		}
		if err != nil {
			t.Errorf("write: %v", err)
		}

		nn := <-c
		if nn != n {
			t.Errorf("wrote %d, read got %d", n, nn)
		}
		t.Logf("wrote %d, read got %d", n, nn)
	}

	w.Close()
	nn := <-c
	if nn != 0 {
		t.Errorf("final read got %d", nn)
	}
}

func TestPipe3(t *testing.T) {
	c := make(chan pipeReturn)
	r, w := Pipe()
	var wdata = make([]byte, 128)
	for i := 0; i < len(wdata); i++ {
		wdata[i] = byte(i)
	}

	go writer(w, wdata, c)

	var rdata = make([]byte, 1024)
	tot := 0
	for n := 1; n <= 256; n *= 2 {
		nn, err := r.Read(rdata[tot : tot+n])
		if err != nil && err != io.EOF {
			t.Fatalf("read: %v", err)
		}
		t.Logf("read %d bytes", nn)

		expect := n
		if n == 128 {
			expect = 1
		} else if n == 256 {
			expect = 0
			if err != io.EOF {
				t.Fatalf("read at end: %v", err)
			}
		}
		if nn != expect {
			t.Fatalf("read %d, expected %d, got %d", n, expect, nn)
		}
		tot += nn
	}

	pr := <-c
	if pr.n != 128 || pr.err != nil {
		t.Fatalf("write 128: %d, %v", pr.n, pr.err)
	}
	if tot != 128 {
		t.Fatalf("total read %d != 128", tot)
	}
	for i := 0; i < 128; i++ {
		if rdata[i] != byte(i) {
			t.Fatalf("rdata[%d] = %d", i, rdata[i])
		}
	}
}

func checkWrite(t *testing.T, w io.Writer, data []byte, c chan int) {
	n, err := w.Write(data)
	if err != nil {
		t.Errorf("write: %v", err)
	}
	if n != len(data) {
		t.Errorf("short write: %d != %d", n, len(data))
	}
	c <- 0
}

func reader(t *testing.T, r io.Reader, c chan int) {
	var buf = make([]byte, 64)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			c <- 0
			break
		}
		if err != nil {
			t.Errorf("read: %v", err)
		}
		c <- n
	}
}

func writer(w io.WriteCloser, buf []byte, c chan pipeReturn) {
	n, err := w.Write(buf)
	w.Close()
	c <- pipeReturn{n, err}
}
