// Command qualithm-mcp is the operator MCP server for the Qualithm platform
// management API: it exposes the management API as agent-native MCP tools
// over stdio, authenticated with a member API token.
//
// The same operator client backs both this server and the qualithm CLI, so the
// agent and human surfaces never diverge.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/qualithm/operator-mcp/internal/server"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "qualithm-mcp: "+err.Error())
		os.Exit(1)
	}
}

// run parses args and serves MCP over stdio. stdout receives --version and
// --help output; once serving, os.Stdout is the MCP transport.
func run(args []string, stdout io.Writer) error {
	var (
		showHelp    bool
		showVersion bool
		token       string
		baseURL     string
	)
	fs := flag.NewFlagSet("qualithm-mcp", flag.ContinueOnError)
	fs.BoolVar(&showHelp, "help", false, "show usage and exit")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.StringVar(&token, "token", os.Getenv("QUALITHM_API_TOKEN"), "member API token (or set QUALITHM_API_TOKEN)")
	fs.StringVar(&baseURL, "url", os.Getenv("QUALITHM_API_URL"), "management API base URL (or set QUALITHM_API_URL)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if showVersion {
		_, _ = fmt.Fprintln(stdout, versionString())
		return nil
	}
	if showHelp {
		fs.SetOutput(stdout)
		_, _ = fmt.Fprintf(stdout, "qualithm-mcp %s — operator MCP server (stdio)\n\n", resolvedVersion())
		fs.Usage()
		return nil
	}

	// Logs must go to stderr: stdout is the MCP transport.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	srv, err := server.New(server.Config{Token: token, BaseURL: baseURL, Version: resolvedVersion()})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("qualithm-mcp starting", "version", resolvedVersion(), "commit", commit, "transport", "stdio")
	return srv.Run(ctx, resolvedVersion())
}
