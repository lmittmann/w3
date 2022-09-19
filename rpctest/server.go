// Package rpctest provides utilities for testing RPC methods.
package rpctest

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

// Server is a fake RPC endpoint that responds only to a single requests that
// is definded in a golden-file.
//
// Request golden-files have the following format to define a single request
// and the corresponding response:
//
//	// Comments and empty lines will be ignored.
//	// Request starts with ">".
//	> {"jsonrpc":"2.0","id":1,"method":"eth_chainId"}
//	// Response starts with "<".
//	< {"jsonrpc":"2.0","id":1,"result":"0x1"}
type Server struct {
	t *testing.T

	reader   io.Reader
	readOnce sync.Once
	in       []byte
	out      []byte

	httptestSrv *httptest.Server
}

// NewServer returns a new instance of Server that serves the golden-file from
// Reader r.
func NewServer(t *testing.T, r io.Reader) *Server {
	srv := &Server{t: t, reader: r}
	httptestSrv := httptest.NewServer(srv)

	srv.httptestSrv = httptestSrv
	return srv
}

// NewFileServer returns a new instance of Server that serves the golden-file
// from the given filename.
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
	body, err := io.ReadAll(r.Body)
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

// URL returns the servers RPC endpoint url.
func (srv *Server) URL() string {
	return srv.httptestSrv.URL
}

// Close shuts down the server.
func (srv *Server) Close() {
	srv.httptestSrv.Close()
}

func (srv *Server) readGolden() {
	if rc, ok := srv.reader.(io.ReadCloser); ok {
		defer rc.Close()
	}

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
