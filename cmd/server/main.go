package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/MuhammadBilal64/go-scale/internal/node"
	"github.com/MuhammadBilal64/go-scale/internal/persistence"
	"github.com/MuhammadBilal64/go-scale/internal/storage"
)

const compactEvery = 100

type Server struct {
	store      *storage.Store
	self       node.Node
	membership *node.Membership
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

type clusterResponse struct {
	Self  node.Node         `json:"self"`
	Peers []node.PeerStatus `json:"peers"`
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

func (s *Server) nodeInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.self)
}

func (s *Server) clusterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, clusterResponse{
		Self:  s.self,
		Peers: s.membership.Snapshot(),
	})
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
	id := flag.String("id", "node-1", "node identity")
	addr := flag.String("addr", ":8085", "HTTP listen address")
	peersFlag := flag.String("peers", "", "comma-separated peers as id@host:port")
	dataDir := flag.String("data-dir", ".", "directory for WAL and snapshot files")
	flag.Parse()

	peers, err := node.ParsePeers(*peersFlag)
	if err != nil {
		log.Fatalf("invalid -peers: %v", err)
	}
	if err := node.ValidateConfig(*id, peers); err != nil {
		log.Fatalf("invalid node config: %v", err)
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	walPath := filepath.Join(*dataDir, "data.wal")
	snapshotPath := filepath.Join(*dataDir, "snapshot.json")

	engine := persistence.NewEngine(walPath, snapshotPath, compactEvery)
	store := storage.NewStore(engine)
	if err := store.Load(); err != nil {
		log.Fatalf("failed to load store: %v", err)
	}

	self := node.Node{ID: *id, Addr: *addr}
	membership := node.NewMembership(peers)

	server := Server{
		store:      store,
		self:       self,
		membership: membership,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/kv/{key}", server.kvHandler)
	mux.HandleFunc("/node/info", server.nodeInfoHandler)
	mux.HandleFunc("/cluster", server.clusterHandler)

	httpServer := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	hbCtx, hbCancel := context.WithCancel(context.Background())
	defer hbCancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		log.Printf("node %s listening on %s", self.ID, *addr)
		err := httpServer.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Start heartbeats only after the listener is bound.
	go membership.RunHeartbeats(hbCtx, 2*time.Second, node.CheckHealth)

	<-quit

	hbCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

// Example two-node run:
//
//	go run ./cmd/server -id node-1 -addr :8085 -peers node-2@127.0.0.1:8086 -data-dir ./data/node1
//	go run ./cmd/server -id node-2 -addr :8086 -peers node-1@127.0.0.1:8085 -data-dir ./data/node2
