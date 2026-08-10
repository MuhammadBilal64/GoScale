package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/MuhammadBilal64/go-scale/internal/persistence"
	"github.com/MuhammadBilal64/go-scale/internal/storage"
)

const (
	walFilePath      = "data.wal"
	snapshotFilePath = "snapshot.json"
	compactEvery     = 100
)

type Server struct {
	store *storage.Store
}
type SetRequest struct {
	Value string `json:"value"`
}
type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func (s *Server) kvHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	switch r.Method {
	case http.MethodGet:
		value, exists := s.store.Get(key)
		if !exists {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "key not found"})
			return
		}
		response := GetResponse{
			Key:   key,
			Value: value,
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPut:
		var req SetRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid Json"})
			return
		}
		if err := s.store.Set(key, req.Value); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to store value"})
			return
		}

		response := map[string]string{
			"message": "value stored",
		}

		writeJSON(w, http.StatusOK, response)
	case http.MethodDelete:
		if err := s.store.Delete(key); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to delete value"})
			return
		}

		response := map[string]string{
			"message": "value deleted",
		}

		writeJSON(w, http.StatusOK, response)
	default:
		writeJSON(
			w,
			http.StatusMethodNotAllowed,
			ErrorResponse{
				Error: "method not allowed",
			},
		)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func main() {
	engine := persistence.NewEngine(walFilePath, snapshotFilePath, compactEvery)
	store := storage.NewStore(engine)
	if err := store.Load(); err != nil {
		log.Fatalf("failed to load store: %v", err)
	}

	server := Server{
		store: store,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/kv/{key}", server.kvHandler)
	httpServer := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

}

//
// go run ./cmd/server

//
