package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type InfoResponse struct {
	Service   string `json:"service"`
	PRNumber  string `json:"pr_number"`
	CommitSHA string `json:"commit_sha"`
	Hostname  string `json:"hostname"`
	Timestamp string `json:"timestamp"`
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	port := getEnv("PORT", "8080")
	prNumber := getEnv("PR_NUMBER", "local")
	commitSHA := getEnv("COMMIT_SHA", "dev")
	hostname, _ := os.Hostname()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, InfoResponse{
			Service:   "ephemeral-pr-demo",
			PRNumber:  prNumber,
			CommitSHA: commitSHA,
			Hostname:  hostname,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (pr=%s, sha=%s)", port, prNumber, commitSHA)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("server stopped cleanly")
}