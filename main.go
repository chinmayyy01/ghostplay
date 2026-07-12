package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	port:= getEnv("PORT", "8080")
	http.HandleFunc("/healthz", healthzHandler)
	fmt.Printf("GhostPlay starting on :%s\n", port)
	fmt.Printf("Try: curl localhost:%s/healthz\n", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}