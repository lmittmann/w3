// Package rpctest provides utilities for RPC testing.
package rpctest

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

type Server struct {
	t *testing.T

	reader   io.Reader
	readOnce sync.Once
	in       []byte
	out      []byte

	httptestSrv *httptest.Server
}

func NewServer(t *testing.T, r io.Reader) *Server {
	srv := &Server{t: t, reader: r}
	httptestSrv := httptest.NewServer(srv)

	srv.httptestSrv = httptestSrv
	return srv
}

func NewFileServer(t *testing.T, filename string) *Server {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	return NewServer(t, f)
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.readOnce.Do(srv.readGolden)

	// read body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.t.Fatalf("Failed to read body: %v", err)
	}

	// check body
	if !bytes.Equal(srv.in, body) {
		srv.t.Fatalf("Invalid request body (-want, +got)\n-%s\n+%s", srv.in, body)
	}

	// respond
	w.Header().Set("Content-Type", "application/json")
	w.Write(srv.out)
}

func (srv *Server) URL() string {
	return srv.httptestSrv.URL
}

func (srv *Server) Close() {
	srv.httptestSrv.Close()
	if rc, ok := srv.reader.(io.ReadCloser); ok {
		rc.Close()
	}
}

func (srv *Server) readGolden() {
	scan := bufio.NewScanner(srv.reader)
	for scan.Scan() {
		line := scan.Bytes()
		if len(line) <= 0 {
			continue // skip empty lines
		}

		switch line[0] {
		case '>':
			trimedLine := bytes.Trim(line, "> ")
			srv.in = make([]byte, len(trimedLine))
			copy(srv.in, trimedLine)
		case '<':
			trimedLine := bytes.Trim(line, "< ")
			srv.out = make([]byte, len(trimedLine))
			copy(srv.out, trimedLine)
		case '/': // ignore lines starting with "/"
		default:
			srv.t.Fatalf("Invalid line %q", scan.Text())
		}
	}
	if err := scan.Err(); err != nil {
		srv.t.Fatalf("Failed to scan file: %v", err)
	}
}
