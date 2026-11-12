package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func call(tool, path string) Call {
	value := Call{}
	value.Function.Name = tool
	value.Function.Arguments, _ = json.Marshal(map[string]string{"path": path})
	return value
}

func TestTraceExtractsResultsWithoutCountingRepeatedHistoryTwice(t *testing.T) {
	var trace bytes.Buffer
	encoder := json.NewEncoder(&trace)
	history := []Message{{Role: "system", Content: "Test"}, {Role: "user", Content: "Read a note."}}
	request := func() {
		t.Helper()
		if err := encoder.Encode(map[string]any{
			"kind": "request", "request": Request{Model: "test-model", Messages: history},
		}); err != nil {
			t.Fatal(err)
		}
	}
	response := func(message Message) {
		t.Helper()
		if err := encoder.Encode(map[string]any{
			"kind": "response", "response": map[string]any{"done": true, "message": message},
		}); err != nil {
			t.Fatal(err)
		}
	}
	request()
	missing := Message{Role: "assistant", ToolCalls: []Call{call("read_note", "draft.txt")}}
	response(missing)
	history = append(history, missing, Message{
		Role: "tool", ToolName: "read_note", Content: `{"error":"file is absent"}`,
	})
	request()
	read := Message{Role: "assistant", ToolCalls: []Call{call("read_note", "notes.txt")}}
	response(read)
	history = append(history, read, Message{
		Role: "tool", ToolName: "read_note", Content: `{"path":"notes.txt","content":"Paper Kite."}`,
	})
	request()
	response(Message{Role: "assistant", Content: "The project is Paper Kite."})
	evidence, err := readTrace(&trace)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.ModelRequests != 3 || len(evidence.RequestedCalls) != 2 || len(evidence.ToolResults) != 2 {
		t.Fatalf("incorrect trace counts: %+v", evidence)
	}
	if !evidence.ToolResults[0].Error || evidence.ToolResults[1].Error {
		t.Fatal("tool errors were not preserved")
	}
	if evidence.Answer != "The project is Paper Kite." || evidence.InitialRequest.Model != "test-model" {
		t.Fatal("lost answer or initial request")
	}
}

func TestTraceRejectsBrokenEvidence(t *testing.T) {
	for _, trace := range []string{
		`{`,
		`{"kind":"unknown"}`,
		`{"kind":"response","response":{"done":true,"message":{"role":"assistant","content":"Guess"}}}`,
		`{"kind":"request","request":{"messages":[{"role":"tool","tool_name":"read_note","content":"{}"}]}}`,
		`{"kind":"request","request":{"messages":[{"role":"assistant","tool_calls":[{"function":{"name":"read_note","arguments":{}}}]}]}}`,
	} {
		if _, err := readTrace(strings.NewReader(trace)); err == nil {
			t.Fatalf("broken trace was accepted: %s", trace)
		}
	}
}

func TestActionCheckRequiresResultsAndExpectedOutcome(t *testing.T) {
	expected := []Action{{Tool: "read_note", Path: "notes.txt"}}
	failed := ToolResult{Call: call("read_note", "notes.txt"), Result: `{"error":"refused"}`, Error: true}
	if *checkActions(expected, nil, true) || *checkActions(expected, []ToolResult{failed}, true) {
		t.Fatal("request or failed read was counted as a successful read")
	}
	success := ToolResult{Call: call("read_note", "notes.txt"), Result: `{"content":"Paper Kite."}`}
	if !*checkActions(expected, []ToolResult{success}, true) {
		t.Fatal("successful read was missed")
	}
	if *checkActions(expected, []ToolResult{success}, false) {
		t.Fatal("incomplete run passed the action check")
	}
	if checkActions(nil, nil, true) != nil {
		t.Fatal("inapplicable action check should not produce a pass")
	}
}

func TestTermScreenIsNotASemanticJudge(t *testing.T) {
	if !*checkTerms([]string{"Go"}, "It uses GO.", true) {
		t.Fatal("case-insensitive answer was rejected")
	}
	if *checkTerms([]string{"Go"}, "The work is ongoing.", true) {
		t.Fatal("substring was mistaken for a word")
	}
	if !*checkTerms([]string{"Go"}, "It does not use Go.", true) {
		t.Fatal("this lexical screen intentionally cannot understand negation")
	}
	if *checkTerms([]string{"Paper Kite"}, "Paper Kite.", false) {
		t.Fatal("incomplete run passed the term check")
	}
	if checkTerms(nil, "", true) != nil {
		t.Fatal("unscored answers should remain unscored")
	}
}

func TestSummaryKeepsFailuresAndSeparateDenominators(t *testing.T) {
	yes, no := true, false
	summary := summarize([]Sample{
		{Variant: "baseline", Completed: true, RequiredActionsMet: &yes, AnswerTermsMet: &yes, DurationMS: 10, WorkspaceUnchanged: true},
		{Variant: "baseline", Completed: false, RequiredActionsMet: &no, AnswerTermsMet: &no, DurationMS: 100},
		{Variant: "baseline", Completed: true, DurationMS: 20, WorkspaceUnchanged: true},
		{Variant: "revised", Completed: true, RequiredActionsMet: &yes},
	}, "baseline")
	if summary.Total != 3 || summary.Completed != 2 || summary.RequiredActions != (Metric{Passed: 1, Applicable: 2}) ||
		summary.AnswerTerms != (Metric{Passed: 1, Applicable: 2}) {
		t.Fatalf("failures or denominators were hidden: %+v", summary)
	}
	if summary.MedianMS != 20 || summary.P95MS != 100 || summary.Unchanged != 2 {
		t.Fatalf("incorrect latency or filesystem summary: %+v", summary)
	}
}

func TestWorkspaceObservationDetectsChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"notes.txt": "before"}
	ok, err := unchanged(root, expected)
	if err != nil || !ok {
		t.Fatalf("unchanged workspace failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	ok, err = unchanged(root, expected)
	if err != nil || ok {
		t.Fatalf("changed workspace was not detected: %v", err)
	}
}

func TestSubprocessOutputLimitCannotBeBypassedByCopy(t *testing.T) {
	var output boundedBuffer
	_, err := io.Copy(&output, strings.NewReader(strings.Repeat("x", 128*1024+1)))
	if err == nil || len(output.String()) > 128*1024 {
		t.Fatal("oversized subprocess output was accepted")
	}
}

func TestFixtureExpectationsAreWellFormed(t *testing.T) {
	var cases []Case
	if err := json.Unmarshal(caseData, &cases); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, task := range cases {
		if task.ID == "" || seen[task.ID] || task.Prompt == "" || task.Review == "" {
			t.Fatalf("invalid evaluation case: %+v", task)
		}
		seen[task.ID] = true
		for name, content := range task.Notes {
			if (name != "notes.txt" && name != "project.txt") || len(content) > 2048 || strings.Contains(content, privateMarker) {
				t.Fatalf("invalid fixture in %s", task.ID)
			}
		}
		for _, action := range task.RequiredActions {
			if action.Tool != "read_note" && action.Tool != "list_notes" {
				t.Fatalf("unsupported expected action in %s", task.ID)
			}
		}
	}
}
