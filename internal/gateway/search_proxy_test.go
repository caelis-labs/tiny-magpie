package gateway

import (
	"bufio"

	"io"
	"net"
	"net/http"

	"slices"

	"sync"
	"testing"
)

// connectProxy is an HTTP proxy that notes the host:port of each CONNECT
// and tunnels it to where dial sends that name.
type connectProxy struct {
	ln   net.Listener
	dial func(addr string) (net.Conn, error)
	mu   sync.Mutex
	got  []string
}

func newConnectProxy(t *testing.T, dial func(addr string) (net.Conn, error)) *connectProxy {
	t.Helper()
	var ln net.Listener
	for range 20 { // a free port, never one of magpie's own (3425-3437)
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		if p := l.Addr().(*net.TCPAddr).Port; p >= 3425 && p <= 3437 {
			l.Close()
			continue
		}
		ln = l
		break
	}
	if ln == nil {
		t.Fatal("no free port")
	}
	p := &connectProxy{ln: ln, dial: dial}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go p.serve(c)
		}
	}()
	return p
}

func (p *connectProxy) URL() string { return "http://" + p.ln.Addr().String() }

func (p *connectProxy) serve(c net.Conn) {
	defer c.Close()
	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil || req.Method != http.MethodConnect {
		io.WriteString(c, "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")
		return
	}
	p.mu.Lock()
	p.got = append(p.got, req.Host)
	p.mu.Unlock()
	up, err := p.dial(req.Host)
	if err != nil {
		io.WriteString(c, "HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n")
		return
	}
	defer up.Close()
	io.WriteString(c, "HTTP/1.1 200 Connection Established\r\n\r\n")
	done := make(chan struct{}, 2)
	go func() { io.Copy(up, br); done <- struct{}{} }()
	go func() { io.Copy(c, up); done <- struct{}{} }()
	<-done
}

func (p *connectProxy) connects() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.got)
}

func (p *connectProxy) reset() {
	p.mu.Lock()
	p.got = nil
	p.mu.Unlock()
}
