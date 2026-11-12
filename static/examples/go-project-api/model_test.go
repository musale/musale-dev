package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestModelHistoryAndTraceContainOnlyFilteredResultsAndSafeErrors(t *testing.T) {
	for _, upstreamFailure := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "upstream error"}[upstreamFailure], func(t *testing.T) {
			api := projectHandler(testKey())
			if upstreamFailure {
				api = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, testKey()+" SYNTHETIC-INTERNAL diagnostic", 500)
				})
			}
			tool := testTool(t, api)
			var trace, log bytes.Buffer
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if strings.Contains(string(body), testKey()) || strings.Contains(string(body), "SYNTHETIC-INTERNAL") ||
					r.Header.Get("X-API-Key") != "" {
					t.Error("model request disclosed application-only data")
				}
				var request ChatRequest
				if err := json.Unmarshal(body, &request); err != nil {
					t.Error(err)
				}
				reply := ChatResponse{Done: true, Message: Message{Role: "assistant"}}
				switch requests {
				case 1:
					if request.Stream || request.Think || !json.Valid(request.Tools) {
						t.Error("unexpected model configuration")
					}
					reply.Message.ToolCalls = []ToolCall{projectCall(`{"id":"night-heron"}`)}
				case 2:
					last := request.Messages[len(request.Messages)-1]
					if last.Role != "tool" || last.ToolName != "get_project" || !strings.Contains(last.Content, "forbidden") {
						t.Error("missing authorization error in history")
					}
					reply.Message.ToolCalls = []ToolCall{projectCall(`{"id":"paper-kite"}`)}
				default:
					last := request.Messages[len(request.Messages)-1]
					expected := "Paper Kite"
					if upstreamFailure {
						expected = "upstream_error"
					}
					if last.Role != "tool" || !strings.Contains(last.Content, expected) {
						t.Errorf("missing result in history: %s", last.Content)
					}
					reply.Message.Content = "The lookup finished."
				}
				json.NewEncoder(w).Encode(reply)
			}))
			defer server.Close()
			client := &ModelClient{endpoint: server.URL, client: server.Client(), trace: json.NewEncoder(&trace)}
			answer, err := turn(context.Background(), client, tool, "test-model", "Look up the project.", &log)
			if err != nil || answer != "The lookup finished." || requests != 3 {
				t.Fatalf("unexpected turn result: %q, %v, %d", answer, err, requests)
			}
			for _, text := range []string{trace.String(), log.String()} {
				if strings.Contains(text, testKey()) || strings.Contains(text, "SYNTHETIC-INTERNAL") {
					t.Fatal("trace or log disclosed application-only data")
				}
			}
			lines := strings.Split(strings.TrimSpace(trace.String()), "\n")
			if len(lines) != 6 {
				t.Fatalf("wanted six trace records, got %d", len(lines))
			}
			for _, line := range lines {
				if !json.Valid([]byte(line)) {
					t.Fatal("invalid JSONL record")
				}
			}
		})
	}
}

func TestModelBudgets(t *testing.T) {
	for _, batch := range []int{1, 2, maxToolCalls + 1} {
		t.Run(strings.Repeat("call ", batch), func(t *testing.T) {
			var apiRequests atomic.Int32
			handler := projectHandler(testKey())
			tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiRequests.Add(1)
				handler.ServeHTTP(w, r)
			}))
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				calls := make([]ToolCall, batch)
				for i := range calls {
					calls[i] = projectCall(`{"id":"paper-kite"}`)
				}
				json.NewEncoder(w).Encode(ChatResponse{Done: true, Message: Message{Role: "assistant", ToolCalls: calls}})
			}))
			defer server.Close()
			_, err := turn(context.Background(), &ModelClient{endpoint: server.URL, client: server.Client()},
				tool, "test", "Continue.", io.Discard)
			if err == nil || !strings.Contains(err.Error(), "budget") {
				t.Fatalf("missing budget error: %v", err)
			}
			expected := maxRequests
			if batch == 2 {
				expected = 5
			} else if batch > maxToolCalls {
				expected = 1
			}
			if requests != expected {
				t.Fatalf("unexpected model request count: %d", requests)
			}
			if int(apiRequests.Load()) != (expected-1)*batch {
				t.Fatal("the over-budget batch reached the project API")
			}
		})
	}
	_, err := turn(context.Background(), nil, nil, "test", strings.Repeat("x", maxHistoryBytes+1), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "byte budget") {
		t.Fatalf("oversized history accepted: %v", err)
	}
}

func TestModelRejectsInvalidResponsesWithoutRawDiagnostics(t *testing.T) {
	for _, body := range []string{
		`{`, `{} {}`, `{"done":false}`, `{"done":true,"done_reason":"length","message":{"role":"assistant"}}`,
		`{"done":true,"message":{"role":"system","content":"unexpected"}}`,
		`{"error":"` + testKey() + `"}`, strings.Repeat(" ", maxResponseBytes+1),
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, body)
		}))
		client := &ModelClient{endpoint: server.URL, client: server.Client()}
		_, err := client.complete(context.Background(), ChatRequest{})
		server.Close()
		if err == nil || strings.Contains(err.Error(), testKey()) {
			t.Fatalf("invalid or disclosed model response: %v", err)
		}
	}
}

func TestMissingKeyFailsBeforeStartingModel(t *testing.T) {
	var output, log bytes.Buffer
	err := run(context.Background(), []string{"-prompt", "Get Paper Kite."},
		func(string) string { return "" }, &output, &log)
	if err == nil || !strings.Contains(err.Error(), "PROJECT_API_KEY") || output.Len() != 0 {
		t.Fatalf("missing-key configuration was not rejected: %v", err)
	}
}
