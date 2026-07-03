// Command qualithm-mcp is the operator MCP server for the qualithm platform
// management API: it exposes the provisioning surface (authorities,
// enrollments, credentials, devices, api-tokens) as agent-native MCP tools over
// stdio, authenticated with a member API token.
//
// The same operator client backs both this server and the qualithm CLI, so the
// agent and human surfaces never diverge.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/qualithm/operator-mcp/internal/server"
)

// version is overridden at release time via -ldflags.
var version = "dev"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "qualithm-mcp: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	var (
		showHelp bool
		token    string
		baseURL  string
	)
	fs := flag.NewFlagSet("qualithm-mcp", flag.ContinueOnError)
	fs.BoolVar(&showHelp, "help", false, "show usage and exit")
	fs.StringVar(&token, "token", os.Getenv("QUALITHM_API_TOKEN"), "member API token (or set QUALITHM_API_TOKEN)")
	fs.StringVar(&baseURL, "url", os.Getenv("QUALITHM_API_URL"), "management API base URL (or set QUALITHM_API_URL)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	if showHelp {
		fs.SetOutput(os.Stdout)
		_, _ = fmt.Fprintf(os.Stdout, "qualithm-mcp %s — operator MCP server (stdio)\n\n", version)
		fs.Usage()
		return nil
	}

	// Logs must go to stderr: stdout is the MCP transport.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	srv, err := server.New(server.Config{Token: token, BaseURL: baseURL})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("qualithm-mcp starting", "version", version, "transport", "stdio")
	return srv.Run(ctx, version)
}
