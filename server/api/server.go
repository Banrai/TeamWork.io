// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
)

type Server struct {
	mux       *http.ServeMux
	s         *http.Server
	Logger    *log.Logger
	Transport string
}

type SimpleMessage struct {
	Ack string `json:"msg"`
	Err string `json:"err,omitempty"`
}

const (
	INVALID_REQUEST   = "Invalid Request"
	INVALID_SESSION   = "Session is expired or invalid"
	MISSING_PARAMETER = "Missing required parameter"
)

var (
	Srv                      *Server
	DefaultServerReadTimeout = 30 // in seconds
	DefaultServerTransport   = "tcp"
)

func GenerateSimpleMessage(msg string, errorMsg string) string {
	ack := new(SimpleMessage)
	ack.Ack = msg
	ack.Err = errorMsg
	reply, _ := json.Marshal(ack)
	return string(reply)
}

func Respond(mediaType string, charset string, fn func(w http.ResponseWriter, r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=%s", mediaType, charset))
		data := fn(w, r)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

func RequestServer(host, transport string, port, timeout int, statics map[string]http.Handler, handlers map[string]func(http.ResponseWriter, *http.Request)) {
	mux := http.NewServeMux()
	for pattern, staticHandler := range statics {
		mux.Handle(pattern, staticHandler)
	}
	for pattern, handler := range handlers {
		mux.Handle(pattern, http.HandlerFunc(handler))
	}
	s := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Handler:     mux,
		ReadTimeout: time.Duration(timeout) * time.Second, // to prevent abuse of "keep-alive" requests by clients
	}
	Srv = &Server{
		mux:       mux,
		s:         s,
		Logger:    log.New(os.Stdout, "", log.Ldate|log.Ltime),
		Transport: transport,
	}

	// create a listener for the incoming FastCGI requests
	listener, err := net.Listen(Srv.Transport, Srv.s.Addr)
	if err != nil {
		Srv.Logger.Fatal(err)
	}
	fcgi.Serve(listener, Srv.mux)
}
