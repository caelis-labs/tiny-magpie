package gui

import (
	"io"
	"net"
	"net/http"
	"os"
)

func fakeOlderGateway(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(3)
	}
	http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"name":"magpie","version":"0.1.550","models":3}`)
	}))
}
