package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
)

var (
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrIPNotFound         = errors.New("IP not found")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "serve:", err)
		os.Exit(1)
	}
}

func run() error {
	port := flag.String("port", "80", "TCP port for the server to listen on")
	flag.Parse()
	if flag.NArg() != 1 {
		return ErrNotEnoughArguments
	}

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	name := flag.Arg(0)
	info, err := getFileStat(name)
	if err != nil {
		return err
	}

	switch mode := info.Mode(); {
	case mode.IsRegular():
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

			http.ServeFile(w, r, name)
		})
	case mode.IsDir():
		mux.Handle("/", http.FileServer(http.Dir(name)))
	default:
		return fmt.Errorf("file mode '%s' not supported", mode)
	}

	ip, err := getLocalIP()
	if err != nil {
		return err
	}

	fmt.Printf("Local http://localhost:%s\n", *port)
	fmt.Printf("Network http://%s:%s\n", ip, *port)

	return srv.ListenAndServe()
}

func getFileStat(name string) (fs.FileInfo, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Stat()
}

func getLocalIP() (net.IP, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}

		return ip, nil
	}

	return nil, ErrIPNotFound
}
