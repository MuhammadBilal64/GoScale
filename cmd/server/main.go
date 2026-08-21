package main

import (
	"net/http"

	"github.com/MuhammadBilal64/go-scale/internal/storage"

	"encoding/json"
)

type Server struct {
	store *storage.Store
}
type SetRequest struct {
	Value string `json:"value"`
}

func (s *Server) kvHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	switch r.Method {
	case http.MethodGet:
		value, exists := s.store.Get(key)
		if !exists {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	case http.MethodPut:
		var req SetRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		s.store.Set(key, req.Value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Value stored"))

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
	store := storage.NewStore()
	server := Server{
		store: store,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/kv/{key}", server.kvHandler)

	err := http.ListenAndServe(":8085", mux)

	if err != nil {
		panic(err)
	}
}

//
// go run ./cmd/server

//
