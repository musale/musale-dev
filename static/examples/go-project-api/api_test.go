package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testKey() string { return strings.Repeat("demo-only-", 4) }

func projectCall(arguments string) ToolCall {
	call := ToolCall{}
	call.Function.Name = "get_project"
	call.Function.Arguments = json.RawMessage(arguments)
	return call
}

func testTool(t *testing.T, handler http.Handler) *ProjectTool {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	tool := newProjectTool(testKey())
	tool.endpoint = server.URL
	t.Cleanup(tool.client.CloseIdleConnections)
	return tool
}

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var detail *ToolError
	if !errors.As(err, &detail) || detail.Code != code {
		t.Fatalf("wanted %q, got %v", code, err)
	}
	if strings.Contains(err.Error(), testKey()) {
		t.Fatal("error disclosed the application key")
	}
}

func TestKeyConfigurationDoesNotEchoInvalidValues(t *testing.T) {
	for _, key := range []string{"", "short", strings.Repeat("x", 129), strings.Repeat("x", 16) + "\n"} {
		_, err := loadKey(func(string) string { return key })
		if err == nil {
			t.Fatal("invalid key was accepted")
		}
		if key != "" && strings.Contains(err.Error(), key) {
			t.Fatal("configuration error echoed the key")
		}
	}
	key, err := loadKey(func(name string) string {
		if name != "PROJECT_API_KEY" {
			t.Fatalf("unexpected environment lookup: %s", name)
		}
		return testKey()
	})
	if err != nil || key != testKey() {
		t.Fatalf("valid key rejected: %v", err)
	}
}

func TestRejectsArgumentsAndUnauthorizedProjectsBeforeHTTP(t *testing.T) {
	var requests atomic.Int32
	tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	for _, raw := range []string{
		`{}`, `null`, `[]`, `{"id":null}`, `{"id":42}`, `{"ID":"paper-kite"}`,
		`{"id":"paper-kite","id":"night-heron"}`, `{"id":"paper-kite"} {}`,
		`{"id":"paper-kite","url":"http://example.test"}`,
		`{"id":"paper-kite","method":"POST"}`,
		`{"id":"paper-kite","user":"administrator"}`,
		`{"id":"paper-kite","api_key":"model-controlled"}`,
		`{"id":"../paper-kite"}`, `{"id":"https://example.test"}`,
		`{"id":"paper-kite?other=1"}`, `{"id":"paper%2fkite"}`,
		strings.Repeat(" ", maxArgumentBytes+1) + `{}`,
		string([]byte{'{', '"', 'i', 'd', '"', ':', '"', 0xff, '"', '}'}),
	} {
		t.Run(raw[:min(len(raw), 70)], func(t *testing.T) {
			_, err := tool.Execute(context.Background(), projectCall(raw))
			requireCode(t, err, "invalid_arguments")
		})
	}
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"night-heron"}`))
	requireCode(t, err, "forbidden")
	tool.principal.ID = ""
	_, err = tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "forbidden")
	for _, unsupported := range []string{"type", "name"} {
		call := projectCall(`{"id":"paper-kite"}`)
		if unsupported == "type" {
			call.Type = "shell"
		} else {
			call.Function.Name = "send_email"
		}
		_, err := tool.Execute(context.Background(), call)
		requireCode(t, err, "unknown_tool")
	}
	if requests.Load() != 0 {
		t.Fatalf("rejected calls made %d HTTP requests", requests.Load())
	}
}

func TestFixedRequestAndFilteredProject(t *testing.T) {
	tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.RequestURI() != "/v1/projects/paper-kite" ||
			r.Header.Get("X-API-Key") != testKey() || r.Header.Get("Accept") != "application/json" {
			t.Error("unexpected API request")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id": "paper-kite", "name": "Paper Kite", "summary": "A local notes application.",
			"internal_note": "SYNTHETIC-INTERNAL " + testKey(),
		})
	}))
	project, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := json.Marshal(project)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `{"id":"paper-kite","name":"Paper Kite","summary":"A local notes application."}` {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestAPIStatusesBecomeSafeErrors(t *testing.T) {
	for status, code := range map[int]string{
		401: "authentication_failed", 403: "upstream_forbidden",
		404: "not_found", 429: "rate_limited", 500: "upstream_error", 503: "upstream_error",
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requests atomic.Int32
			tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.Header().Set("Retry-After", "9999999")
				w.WriteHeader(status)
				io.WriteString(w, testKey()+" SYNTHETIC-INTERNAL raw upstream diagnostics")
			}))
			_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
			requireCode(t, err, code)
			if strings.Contains(err.Error(), "SYNTHETIC-INTERNAL") || strings.Contains(err.Error(), "9999999") {
				t.Fatal("raw upstream data reached the error")
			}
			if requests.Load() != 1 {
				t.Fatal("API request was automatically retried")
			}
		})
	}
}

func TestRedirectIsNotFollowed(t *testing.T) {
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirected.Add(1)
	}))
	defer target.Close()
	tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "upstream_error")
	if redirected.Load() != 0 {
		t.Fatal("redirect target received a request")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestTransportDiagnosticsAreNotDisclosed(t *testing.T) {
	tool := newProjectTool(testKey())
	defer tool.client.CloseIdleConnections()
	tool.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("SYNTHETIC-INTERNAL diagnostic containing " + testKey())
	})
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "unavailable")
	if strings.Contains(err.Error(), "SYNTHETIC-INTERNAL") {
		t.Fatal("raw transport diagnostic was disclosed")
	}
}

func TestInvalidResponseDoesNotReachTheModel(t *testing.T) {
	for _, body := range []string{
		`{"id":"paper-kite","name":"Paper Kite","summary":"x"} {}`,
		`{"id":"night-heron","name":"Wrong project","summary":"x"}`,
		`{"id":"paper-kite","name":"Paper Kite"}`,
		`{"id":"paper-kite","name":42,"summary":"x"}`,
		`{"id":`, strings.Repeat("x", maxAPIBytes+1),
		string([]byte{0xff}),
		`{"id":"paper-kite","name":"Paper Kite","summary":"` + strings.Repeat("x", 1025) + `"}`,
	} {
		tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, body)
		}))
		_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
		requireCode(t, err, "invalid_response")
	}
	tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html>upstream proxy error</html>")
	}))
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "invalid_response")
}

func TestDeadlinesAndCancellation(t *testing.T) {
	tool := testTool(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	tool.client.Timeout = 30 * time.Millisecond
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "timeout")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = tool.Execute(ctx, projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "cancelled")
}

func TestServiceAuthenticationIsNotUserAuthorization(t *testing.T) {
	handler := projectHandler(testKey())
	for _, key := range []string{"", "incorrect-application-key"} {
		request := httptest.NewRequest(http.MethodGet, "/v1/projects/paper-kite", nil)
		request.Header.Set("X-API-Key", key)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("invalid key returned HTTP %d", response.Code)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/projects/night-heron", nil)
	request.Header.Set("X-API-Key", testKey())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Night Heron") {
		t.Fatal("service key should access both synthetic projects")
	}
	tool := testTool(t, handler)
	_, err := tool.Execute(context.Background(), projectCall(`{"id":"night-heron"}`))
	requireCode(t, err, "forbidden")
	tool.key = "incorrect-application-key"
	_, err = tool.Execute(context.Background(), projectCall(`{"id":"paper-kite"}`))
	requireCode(t, err, "authentication_failed")
}
