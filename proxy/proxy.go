package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/chinmayyy01/ghostplay/storage"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func ProxyHandler(targetURL string, store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		target := targetURL + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequest(r.Method, target, bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, "failed to build proxy request", http.StatusInternalServerError)
			return
		}


		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		start := time.Now()
		resp, err := httpClient.Do(proxyReq)
		duration := time.Since(start)
		if err != nil {
			http.Error(w, fmt.Sprintf("upstream error: %v", err), http.StatusBadGateway)
			log.Printf("PROXY ERROR %s %s -> %v", r.Method, r.URL.Path, err)
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "failed to read upstream response", http.StatusInternalServerError)
			return
		}

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
 
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, resp.StatusCode, duration)

		record := storage.NewRecord(
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			r.Header,
			string(bodyBytes),
			resp.StatusCode,
			resp.Header,
			string(respBody),
			duration.Milliseconds(),
		)
		if err := store.Save(record); err != nil {
			log.Printf("RECORD ERROR failed to save session: %v", err)
		}
	}		
}
