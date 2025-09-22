package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/bhochhi/lexy/internal/logx"
	"github.com/bhochhi/lexy/nlu"
	"github.com/bhochhi/lexy/orchestrator"
	"github.com/bhochhi/lexy/registry"
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
	// Use central registry to avoid editing main.go when adding intents
	funcs := registry.New()
	orc := orchestrator.New(store, nlu.NewClient(), funcs, workDir())

	// Structured logger (slog JSON)
	logx.Init("lexy")

	// Local server in DEBUG mode; otherwise Lambda handler
	if os.Getenv("DEBUG") == "true" {
		mux := http.NewServeMux()
		mux.Handle("/chat", logx.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			// add request-scoped clientId to logger
			ctx := r.Context()
			if req.ClientID != "" {
				ctx = logx.WithFields(ctx, map[string]any{"client_id": req.ClientID})
			}
			text, opts, err := orc.Execute(req.ClientID, req.UserText)
			if err != nil {
				logx.LogError(ctx, "execute_error", err)
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
				return
			}
			_ = json.NewEncoder(w).Encode(ChatResponse{ResponseText: text, Options: opts})
		})))

		addr := net.JoinHostPort(getenv("HOST", "127.0.0.1"), getenv("PORT", "8080"))
		log.Printf("listening on %s", addr)
		log.Fatal(http.ListenAndServe(addr, mux))
		return
	}

	// Lambda mode
	lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// Extract headers for context
		// Note: In Lambda, attach fields to logger context for correlation if desired
		if v := req.Headers["x-request-id"]; v != "" {
			ctx = logx.WithFields(ctx, map[string]any{"request_id": v})
		}
		if v := req.Headers["client-id"]; v != "" {
			ctx = logx.WithFields(ctx, map[string]any{"client_id": v})
		}
		if v := req.Headers["conversation-id"]; v != "" {
			ctx = logx.WithFields(ctx, map[string]any{"conversation_id": v})
		}

		// Parse body
		var cr ChatRequest
		if err := json.Unmarshal([]byte(req.Body), &cr); err != nil {
			logx.ErrorContext(ctx, "bad_request", map[string]any{"error": "invalid json"})
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: `{"error":"invalid json"}`}, nil
		}
		if cr.ClientID != "" {
			ctx = logx.WithFields(ctx, map[string]any{"client_id": cr.ClientID})
		}

		text, opts, err := orc.Execute(cr.ClientID, cr.UserText)
		if err != nil {
			logx.ErrorContext(ctx, "execute_error", logx.ErrField(err))
			b, _ := json.Marshal(map[string]any{"error": err.Error()})
			return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: string(b)}, nil
		}
		b, _ := json.Marshal(ChatResponse{ResponseText: text, Options: opts})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(b),
			Headers:    map[string]string{"content-type": "application/json"},
		}, nil
	})
}

// workDir returns the workflow directory path
func workDir() string {
	if v := os.Getenv("WORKFLOW_DIR"); v != "" {
		return v
	}
	return "workflow"
}

// nluAdapter implements orchestrator.NLU using our nlu package stub

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
