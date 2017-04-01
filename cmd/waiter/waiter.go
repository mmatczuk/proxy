package main

import (
	"flag"
	"net/http"
)

func main() {
	flag.Parse()

	for _, addr := range flag.Args() {
		addr := addr
		go func() {
			http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			}))
		}()
	}

	select {}
}
