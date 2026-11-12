package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func serve(ctx context.Context, key string, log io.Writer) error {
	listener, err := net.Listen("tcp", "127.0.0.1:18080")
	if err != nil {
		return errors.New("could not listen on 127.0.0.1:18080; check whether the demo API is already running")
	}
	server := &http.Server{
		Handler: projectHandler(key), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8 * 1024,
	}
	if _, err := fmt.Fprintln(log, "Synthetic project API listening on http://127.0.0.1:18080"); err != nil {
		return errors.Join(err, listener.Close())
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, output, log io.Writer) (err error) {
	flags := flag.NewFlagSet("project-api-tools", flag.ContinueOnError)
	flags.SetOutput(log)
	server := flags.Bool("serve", false, "run the synthetic project API")
	project := flags.String("project", "", "look up a project directly, without the model")
	prompt := flags.String("prompt", "", "one request for the local model")
	model := flags.String("model", "qwen3:0.6b", "local Ollama model")
	trace := flags.String("trace", "", "create a private model JSONL trace")
	if err := flags.Parse(args); err != nil {
		return err
	}
	modes := 0
	for _, selected := range []bool{*server, *project != "", strings.TrimSpace(*prompt) != ""} {
		if selected {
			modes++
		}
	}
	if flags.NArg() != 0 || modes != 1 || (*trace != "" && strings.TrimSpace(*prompt) == "") {
		return errors.New("choose exactly one of -serve, -project ID, or -prompt TEXT; -trace requires -prompt")
	}
	key, err := loadKey(getenv)
	if err != nil {
		return err
	}
	if *server {
		return serve(ctx, key, log)
	}
	tool := newProjectTool(key)
	defer tool.client.CloseIdleConnections()
	if *project != "" {
		var call ToolCall
		call.Function.Name = "get_project"
		call.Function.Arguments, err = json.Marshal(struct {
			ID string `json:"id"`
		}{*project})
		if err != nil {
			return err
		}
		result, err := tool.Execute(ctx, call)
		if err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(result)
	}
	client := &ModelClient{
		endpoint: "http://127.0.0.1:11434/api/chat",
		client:   localClient(2 * time.Minute),
	}
	defer client.client.CloseIdleConnections()
	if *trace != "" {
		file, openErr := os.OpenFile(*trace, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if openErr != nil {
			return errors.New("could not create trace; choose a new path in a private directory")
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(err, errors.New("could not close model trace"))
			}
		}()
		client.trace = json.NewEncoder(file)
	}
	turnCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	answer, err := turn(turnCtx, client, tool, *model, *prompt, log)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Assistant (untrusted text): %q\n", answer)
	return err
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %q\n", err.Error())
		os.Exit(1)
	}
}
