package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func toolCall(name, arguments string) ToolCall {
	call := ToolCall{}
	call.Function.Name = name
	call.Function.Arguments = json.RawMessage(arguments)
	return call
}

func workspace(t *testing.T, text string) *os.Root {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root
}

func TestCompleteStreamsJSONAcrossWrites(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing POST or JSON content type")
		}
		var request ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if !request.Stream || request.Think {
			t.Error("streaming and thinking flags changed")
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprint(w, `{"message":{"role":"assistant","content":"Hel`)
		w.(http.Flusher).Flush()
		fmt.Fprint(w, "lo\"},\"done\":false}\n")
		fmt.Fprint(w, "{\"message\":{\"content\":\" there\"},\"done\":true,\"done_reason\":\"stop\"}\n")
	}))
	defer server.Close()
	var output, trace bytes.Buffer
	client := &Client{URL: server.URL, HTTP: server.Client(), Trace: json.NewEncoder(&trace)}
	reply, err := client.complete(context.Background(), ChatRequest{Stream: true}, func(chunk string) error {
		_, err := output.WriteString(chunk)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply.Content != "Hello there" || output.String() != reply.Content {
		t.Fatalf("unexpected assembled output: %#v %q", reply, output.String())
	}
	lines := strings.Split(strings.TrimSpace(trace.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected request and two response records, got %d", len(lines))
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("invalid JSONL record: %q", line)
		}
	}
}

func TestCompleteReportsFailures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"HTTP error", 404, `{"error":"model missing"}`, "HTTP 404"},
		{"stream error", 200, `{"error":"runner failed"}`, "runner failed"},
		{"missing completion", 200, `{"message":{"content":"partial"},"done":false}`, "before done:true"},
		{"invalid JSON", 200, `{"message":`, "decode response"},
		{"output limit", 200, `{"message":{"content":"partial"},"done":true,"done_reason":"length"}`, "output limit"},
		{"unexpected role", 200, `{"message":{"role":"system"},"done":true}`, "unexpected response role"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()
			client := &Client{URL: server.URL, HTTP: server.Client()}
			_, err := client.complete(context.Background(), ChatRequest{}, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want error containing %q", err, test.want)
			}
		})
	}
}

func TestTurnRetainsAssistantCallAndToolResult(t *testing.T) {
	root := workspace(t, "The demonstration code is BLUE-47.")
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		var request ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if count == 1 {
			json.NewEncoder(w).Encode(ChatResponse{
				Message: Message{Role: "assistant", ToolCalls: []ToolCall{
					toolCall("read_file", `{"path":"notes.txt"}`),
					toolCall("read_file", `{"path":"notes.txt"}`),
				}},
				Done: true,
			})
			return
		}
		if len(request.Messages) != 5 ||
			len(request.Messages[2].ToolCalls) != 2 ||
			request.Messages[3].Role != "tool" ||
			request.Messages[3].ToolName != "read_file" ||
			request.Messages[3].Content != "The demonstration code is BLUE-47." ||
			request.Messages[4].Role != "tool" ||
			request.Messages[4].Content != request.Messages[3].Content {
			t.Errorf("tool history lost: %#v", request.Messages)
		}
		json.NewEncoder(w).Encode(ChatResponse{Message: Message{Role: "assistant", Content: "BLUE-47"}, Done: true})
	}))
	defer server.Close()
	history := []Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: "Read the code."}}
	result, err := turn(context.Background(), &Client{URL: server.URL, HTTP: server.Client()}, "test", history, root, false, nil, io.Discard)
	if err != nil || count != 2 || len(result) != 6 || result[5].Content != "BLUE-47" {
		t.Fatalf("unexpected tool loop result: %v, requests=%d, history=%#v", err, count, result)
	}
}

func TestReadFileRejectsUnauthorizedInputs(t *testing.T) {
	root := workspace(t, "a short note")
	tests := []ToolCall{
		toolCall("shell", `{"path":"notes.txt"}`),
		toolCall("read_file", `{"path":"../notes.txt"}`),
		toolCall("read_file", `{"path":"/etc/passwd"}`),
		toolCall("read_file", `{}`),
		toolCall("read_file", `{"path":"notes.txt","command":"anything"}`),
		toolCall("read_file", `{"path":"notes.txt"} {}`),
		toolCall("read_file", `null`),
	}
	for _, call := range tests {
		if _, err := readFile(root, call); err == nil {
			t.Errorf("accepted unauthorized tool call: %#v", call)
		}
	}
	if _, err := readFile(workspace(t, strings.Repeat("x", maxFileBytes+1)), toolCall("read_file", `{"path":"notes.txt"}`)); err == nil {
		t.Error("accepted an oversized note")
	}
	if _, err := readFile(workspace(t, string([]byte{0xff})), toolCall("read_file", `{"path":"notes.txt"}`)); err == nil {
		t.Error("accepted non-UTF-8 data")
	}
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside.txt"), filepath.Join(dir, "notes.txt")); err != nil {
		t.Fatal(err)
	}
	symlinkRoot, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer symlinkRoot.Close()
	if _, err := readFile(symlinkRoot, toolCall("read_file", `{"path":"notes.txt"}`)); err == nil {
		t.Error("accepted a symlink")
	}
}

func TestTurnLogsToolErrorsAndSendsThemBack(t *testing.T) {
	root := workspace(t, "a short note")
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			json.NewEncoder(w).Encode(ChatResponse{
				Message: Message{Role: "assistant", ToolCalls: []ToolCall{toolCall("read_file", `{"path":"../private.txt"}`)}},
				Done:    true,
			})
			return
		}
		var request ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if len(request.Messages) != 3 || !strings.Contains(request.Messages[2].Content, "only notes.txt is allowed") {
			t.Errorf("missing explicit tool error: %#v", request.Messages)
		}
		json.NewEncoder(w).Encode(ChatResponse{Message: Message{Role: "assistant", Content: "The file is unavailable."}, Done: true})
	}))
	defer server.Close()
	var log bytes.Buffer
	_, err := turn(context.Background(), &Client{URL: server.URL, HTTP: server.Client()}, "test",
		[]Message{{Role: "user", Content: "Read a private file."}}, root, false, nil, &log)
	if err != nil || !strings.Contains(log.String(), "tool error: only notes.txt is allowed") {
		t.Fatalf("got %v with log %q", err, log.String())
	}
}

func TestTurnStopsRepeatedToolCalls(t *testing.T) {
	root := workspace(t, "a short note")
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		json.NewEncoder(w).Encode(ChatResponse{
			Message: Message{Role: "assistant", ToolCalls: []ToolCall{toolCall("read_file", `{"path":"notes.txt"}`)}},
			Done:    true,
		})
	}))
	defer server.Close()
	history := []Message{{Role: "user", Content: "Read the note."}}
	_, err := turn(context.Background(), &Client{URL: server.URL, HTTP: server.Client()}, "test", history, root, false, nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "budget") || count != maxRequests {
		t.Fatalf("got %v after %d requests", err, count)
	}
}
