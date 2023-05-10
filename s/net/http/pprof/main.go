package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "OK")
	})

	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
