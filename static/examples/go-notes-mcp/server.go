package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"example.com/note-tools/notes"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const protocolVersion = "2026-07-28"
const maxCalls = 16

func newServer(policy *notes.Notes) (*mcp.Server, error) {
	server := mcp.NewServer(&mcp.Implementation{Name: "local-notes", Version: "1.0.0"}, &mcp.ServerOptions{
		SupportedProtocolVersions: []string{protocolVersion},
		Capabilities: &mcp.ServerCapabilities{
			Tools: &mcp.ToolCapabilities{},
		},
	})
	var definitions []struct {
		Function struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Parameters  json.RawMessage `json:"parameters"`
		} `json:"function"`
	}
	if err := json.Unmarshal(notes.ToolDefinitions, &definitions); err != nil {
		return nil, fmt.Errorf("decode notes definitions: %w", err)
	}
	var mu sync.Mutex
	calls := 0
	for _, definition := range definitions {
		function := definition.Function
		server.AddTool(&mcp.Tool{
			Name:        function.Name,
			Description: function.Description,
			InputSchema: function.Parameters,
		}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// MCP may dispatch concurrent calls; one Notes value owns one proposal.
			mu.Lock()
			defer mu.Unlock()
			if calls >= maxCalls {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: "tool-call budget exhausted; start a new demonstration process"}},
				}, nil
			}
			calls++
			call := notes.ToolCall{}
			call.Function.Name = function.Name
			call.Function.Arguments = request.Params.Arguments
			if len(call.Function.Arguments) == 0 {
				call.Function.Arguments = json.RawMessage(`{}`)
			}
			text, err := policy.Execute(ctx, call)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				}, nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, nil
		})
	}
	return server, nil
}
