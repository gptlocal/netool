package httptest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

type Server struct {
	URL      string // base URL of form http://ipaddr:port with no trailing slash
	Listener net.Listener

	// EnableHTTP2 controls whether HTTP/2 is enabled
	// on the server. It must be set between calling
	// NewUnstartedServer and calling Server.StartTLS.
	EnableHTTP2 bool

	// TLS is the optional TLS configuration, populated with a new config
	// after TLS is started. If set on an unstarted server before StartTLS
	// is called, existing fields are copied into the new config.
	TLS *tls.Config

	// Config may be changed after calling NewUnstartedServer and
	// before Start or StartTLS.
	Config *http.Server

	// certificate is a parsed version of the TLS config certificate, if present.
	certificate *x509.Certificate

	// wg counts the number of outstanding HTTP requests on this server.
	// Close blocks until all requests are finished.
	wg sync.WaitGroup

	mu     sync.Mutex // guards closed and conns
	closed bool
	conns  map[net.Conn]http.ConnState // except terminal states

	// client is configured for use with the server.
	// Its transport is automatically closed when Close is called.
	client *http.Client
}

func NewServer(handler http.Handler) *Server {
	ts := NewUnstartedServer(handler)
	ts.Start()
	return ts
}

func NewUnstartedServer(handler http.Handler) *Server {
	return &Server{
		Listener: newLocalListener(),
		Config:   &http.Server{Handler: handler},
	}
}

func (s *Server) Start() {
	if s.URL != "" {
		panic("Server already started")
	}
	if s.client == nil {
		s.client = &http.Client{Transport: &http.Transport{}}
	}
	s.URL = "http://" + s.Listener.Addr().String()
	s.wrap()
	s.goServe()
	if serveFlag != "" {
		fmt.Fprintln(os.Stderr, "httptest: serving on", s.URL)
		select {}
	}
}

func newLocalListener() net.Listener {
	if serveFlag != "" {
		l, err := net.Listen("tcp", serveFlag)
		if err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on %v: %v", serveFlag, err))
		}
		return l
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func (s *Server) wrap() {
	oldHook := s.Config.ConnState
	s.Config.ConnState = func(c net.Conn, cs http.ConnState) {
		s.mu.Lock()
		defer s.mu.Unlock()

		switch cs {
		case http.StateNew:
			if _, exists := s.conns[c]; exists {
				panic("invalid state transition")
			}
			if s.conns == nil {
				s.conns = make(map[net.Conn]http.ConnState)
			}
			// Add c to the set of tracked conns and increment it to the waitgroup.
			s.wg.Add(1)
			s.conns[c] = cs
			if s.closed {
				s.closeConn(c)
			}
		case http.StateActive:
			if oldState, ok := s.conns[c]; ok {
				if oldState != http.StateNew && oldState != http.StateIdle {
					panic("invalid state transition")
				}
				s.conns[c] = cs
			}
		case http.StateIdle:
			if oldState, ok := s.conns[c]; ok {
				if oldState != http.StateActive {
					panic("invalid state transition")
				}
				s.conns[c] = cs
			}
			if s.closed {
				s.closeConn(c)
			}
		case http.StateHijacked, http.StateClosed:
			// Remove c from the set of tracked conns and decrement it from the
			// waitgroup, unless it was previously removed.
			if _, ok := s.conns[c]; ok {
				delete(s.conns, c)
				// Keep Close from returning until the user's ConnState hook
				// (if any) finishes.
				defer s.wg.Done()
			}
		}
		if oldHook != nil {
			oldHook(c, cs)
		}
	}
}

func (s *Server) goServe() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.Config.Serve(s.Listener)
	}()
}

func (s *Server) closeConn(c net.Conn) { s.closeConnChan(c, nil) }

func (s *Server) closeConnChan(c net.Conn, done chan<- struct{}) {
	c.Close()
	if done != nil {
		done <- struct{}{}
	}
}
