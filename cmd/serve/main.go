package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	dir := flag.String("dir", "web", "directory of static WASM files")
	flag.Parse()

	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		log.Fatalf("static dir %s not found; build WASM into web/ first (see README)", abs)
	}

	log.Printf("serving %s at http://%s", abs, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(abs))))
}
