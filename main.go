package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

var ErrTooManyArguments = errors.New("too many arguments")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "serve:", err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("o", "0.0.0.0:8080", "TCP address for the server to listen on")
	flag.Parse()
	if flag.NArg() > 1 {
		return ErrTooManyArguments
	}

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	if flag.NArg() == 0 {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Accept-Ranges", "bytes")
			io.Copy(w, bytes.NewReader(stdin))
		})
	} else {
		name := flag.Arg(0)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, name)
		})
	}

	return srv.ListenAndServe()
}
