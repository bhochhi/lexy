package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/bhochhi/lexy/intent/findDeductible"
	"github.com/bhochhi/lexy/nlu"
	"github.com/bhochhi/lexy/orchestrator"
	"github.com/bhochhi/lexy/session"
)

type ChatRequest struct {
	UserText string `json:"userText"`
	ClientID string `json:"clientId"`
}

type ChatResponse struct {
	ResponseText string           `json:"responseText"`
	Options      []session.Option `json:"options"`
}

func main() {
	store := session.NewStore()
	funcs := findDeductible.NewRegistry()
	orc := orchestrator.New(store, nluAdapter{}, funcs, workDir())

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
			return
		}
		text, opts, err := orc.Execute(req.ClientID, req.UserText)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(ChatResponse{ResponseText: text, Options: opts})
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// workDir returns the workflow directory path
func workDir() string {
	if v := os.Getenv("WORKFLOW_DIR"); v != "" {
		return v
	}
	return "workflow"
}

// nluAdapter implements orchestrator.NLU using our nlu package stub

type nluAdapter struct{}

func (n nluAdapter) DetectIntent(text, clientID string) (string, map[string]string, error) {
	return nlu.DetectIntent(text, clientID)
}
