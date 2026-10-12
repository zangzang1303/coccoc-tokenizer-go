package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

type TokenizeRequest struct {
	Text string `json:"text"`
}

type TokenizeResponse struct {
	Original  string             `json:"original"`
	Segmented string             `json:"segmented,omitempty"`
	Tokens    []string           `json:"tokens,omitempty"`
	Details   []coccoc.TokenInfo `json:"details,omitempty"`
	Engine    string             `json:"engine"`
	LatencyMs float64            `json:"latency_ms"`
}

func main() {
	tok, err := coccoc.New(
		coccoc.WithFallbackToPureGo(true),
		coccoc.WithNormalizeLowercase(true),
	)
	if err != nil {
		log.Fatalf("Khởi tạo tokenizer thất bại: %v", err)
	}
	defer tok.Close()

	engineName := "Pure-Go"
	if tok.IsCGO() {
		engineName = "C++ CGO"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy",
			"engine": engineName,
			"is_cgo": tok.IsCGO(),
		})
	})

	http.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
		text := extractText(r)
		if text == "" {
			http.Error(w, `{"error":"Trường 'text' không được rỗng"}`, http.StatusBadRequest)
			return
		}

		start := time.Now()
		seg, err := tok.Segment(text)
		latency := float64(time.Since(start).Microseconds()) / 1000.0

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenizeResponse{
			Original:  text,
			Segmented: seg,
			Engine:    engineName,
			LatencyMs: latency,
		})
	})

	http.HandleFunc("/tokenize", func(w http.ResponseWriter, r *http.Request) {
		text := extractText(r)
		if text == "" {
			http.Error(w, `{"error":"Trường 'text' không được rỗng"}`, http.StatusBadRequest)
			return
		}

		start := time.Now()
		tokens, err := tok.Tokenize(text)
		latency := float64(time.Since(start).Microseconds()) / 1000.0

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenizeResponse{
			Original:  text,
			Tokens:    tokens,
			Engine:    engineName,
			LatencyMs: latency,
		})
	})

	http.HandleFunc("/tokenize_full", func(w http.ResponseWriter, r *http.Request) {
		text := extractText(r)
		if text == "" {
			http.Error(w, `{"error":"Trường 'text' không được rỗng"}`, http.StatusBadRequest)
			return
		}

		start := time.Now()
		details, err := tok.TokenizeFull(text)
		latency := float64(time.Since(start).Microseconds()) / 1000.0

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenizeResponse{
			Original:  text,
			Details:   details,
			Engine:    engineName,
			LatencyMs: latency,
		})
	})

	addr := ":8080"
	fmt.Printf("🚀 Cốc Cốc Tokenizer Microservice đang lắng nghe tại http://localhost%s (%s engine)\n", addr, engineName)
	fmt.Println("👉 GET  /health")
	fmt.Println("👉 GET/POST /segment?text=...")
	fmt.Println("👉 GET/POST /tokenize?text=...")
	fmt.Println("👉 GET/POST /tokenize_full?text=...")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server sập: %v", err)
	}
}

func extractText(r *http.Request) string {
	if r.Method == http.MethodGet {
		return strings.TrimSpace(r.URL.Query().Get("text"))
	}
	var req TokenizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		return strings.TrimSpace(req.Text)
	}
	return ""
}
