package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chinmayyy01/ghostplay/admin"
	"github.com/chinmayyy01/ghostplay/proxy"
	"github.com/chinmayyy01/ghostplay/storage"
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
	port := getEnv("PORT", "8080")
	adminPort := getEnv("ADMIN_PORT", "8081")
	
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		log.Fatal("TARGET_URL env var is required, e.g. TARGET_URL=https://httpbin.org go run .")
	}

	dataFile := getEnv("DATA_FILE", "data/sessions.jsonl")
	store, err := storage.NewStore(dataFile)
	if err != nil {
		log.Fatalf("failed to open data file %s: %v", dataFile, err)
	}

	go func() {
		adminMux := admin.NewMux(store)
		log.Printf("Admin API listening on :%s", adminPort)
		if err := http.ListenAndServe(":"+adminPort, adminMux); err != nil {
			log.Fatal(err)
		}
	}()



	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/", proxy.ProxyHandler(targetURL, store))

	fmt.Printf("GhostPlay proxy on :%s -> %s\n", port, targetURL)
	fmt.Printf("Admin API on :%s\n", adminPort)
	fmt.Printf("Recording sessions to %s\n", dataFile)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}