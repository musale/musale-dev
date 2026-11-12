package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"example.com/note-tools/notes"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func serve(ctx context.Context, workspace string) (err error) {
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return fmt.Errorf("open workspace: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	server, err := newServer(notes.New(root))
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}

func connect(ctx context.Context, command *exec.Cmd) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "notes-demo-client", Version: "1.0.0"}, &mcp.ClientOptions{
		Capabilities: &mcp.ClientCapabilities{},
	})
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, &mcp.ClientSessionOptions{
		ProtocolVersion: protocolVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to notes server: %w", err)
	}
	if session.InitializeResult().ProtocolVersion != protocolVersion {
		closeErr := session.Close()
		return nil, errors.Join(errors.New("server did not select the example's protocol version"), closeErr)
	}
	return session, nil
}

func useServer(ctx context.Context, command *exec.Cmd, tool string, arguments json.RawMessage, output io.Writer) (err error) {
	session, err := connect(ctx, command)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, session.Close()) }()
	encoder := json.NewEncoder(output)
	if tool == "" {
		result, err := session.ListTools(ctx, nil)
		if err != nil {
			return fmt.Errorf("list tools: %w", err)
		}
		return encoder.Encode(result)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: arguments})
	if err != nil {
		return fmt.Errorf("call tool: %w", err)
	}
	if err := encoder.Encode(result); err != nil {
		return err
	}
	if result.IsError {
		return errors.New("tool refused the request; see the result's content")
	}
	return nil
}

func run(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("notes-mcp", flag.ContinueOnError)
	serverMode := flags.Bool("serve", false, "serve MCP on stdin/stdout (no terminal prompts)")
	workspace := flags.String("root", "../go-note-tools/workspace", "synthetic notes workspace")
	tool := flags.String("tool", "", "call this tool; omit to list tool definitions")
	arguments := flags.String("args", "{}", "JSON arguments for the selected tool")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || (*serverMode && (*tool != "" || *arguments != "{}")) {
		return errors.New("use -serve alone with -root, or client mode with optional -tool and -args")
	}
	if *tool == "" && *arguments != "{}" {
		return errors.New("-args requires -tool")
	}
	if *serverMode {
		return serve(ctx, *workspace)
	}
	if len(*arguments) > notes.MaxArgumentBytes || !json.Valid([]byte(*arguments)) {
		return errors.New("-args must be valid JSON of at most 16384 bytes")
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	root, err := filepath.Abs(*workspace)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-serve", "-root", root)
	command.Stderr = os.Stderr
	command.WaitDelay = time.Second
	return useServer(ctx, command, *tool, json.RawMessage(*arguments), output)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %q\n", err.Error())
		os.Exit(1)
	}
}
