package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

// exitCodeConfig marks a missing-configuration exit: the eval never ran, so
// the workflow reports nothing (a standing red issue for an operational
// precondition would be noise); it only notes it in the run summary.
const exitCodeConfig = 2

func main() {
	os.Exit(runMain())
}

func runMain() int {
	var (
		serverBin  string
		apiURL     string
		brokerHost string
		brokerPort int
		token      string
		zone       string
		outPath    string
	)
	fs := flag.NewFlagSet("agent-eval", flag.ContinueOnError)
	fs.StringVar(&serverBin, "server-bin", os.Getenv("AGENT_EVAL_SERVER_BIN"), "path to the qualithm-mcp binary")
	fs.StringVar(&apiURL, "api-url", envOr("QUALITHM_API_URL", "https://api.sandbox.qualithm.com"), "management API base URL")
	fs.StringVar(&brokerHost, "broker-host", envOr("AGENT_EVAL_BROKER_HOST", "gw.sandbox-sg-sin-a.qualithm.com"), "device gateway host")
	fs.IntVar(&brokerPort, "broker-port", 8883, "device gateway TLS port")
	fs.StringVar(&token, "token", os.Getenv("QUALITHM_API_TOKEN"), "member API token for the eval team")
	fs.StringVar(&zone, "zone", os.Getenv("AGENT_EVAL_ZONE"), "device zone for the eval space when the team has none")
	fs.StringVar(&outPath, "out", "", "write the JSON report here (default: stdout only)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return exitCodeConfig
	}

	if serverBin == "" || token == "" {
		_, _ = fmt.Fprintln(os.Stderr, "agent-eval: -server-bin and -token (QUALITHM_API_TOKEN) are required; the eval did not run")
		return exitCodeConfig
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := append(os.Environ(),
		"QUALITHM_API_TOKEN="+token,
		"QUALITHM_API_URL="+apiURL,
	)
	tools, cleanup, err := connectMCPServer(ctx, serverBin, env)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "agent-eval: "+err.Error())
		return 1
	}
	defer cleanup()

	report := Run(ctx, tools,
		&httpProvisioner{baseURL: apiURL, http: &http.Client{Timeout: 30 * time.Second}},
		&mqttPublisher{host: brokerHost, port: brokerPort},
		Config{Zone: zone},
	)

	fmt.Print(report.Render())
	if outPath != "" {
		raw, err := json.MarshalIndent(report, "", "  ")
		if err == nil {
			if werr := os.WriteFile(outPath, append(raw, '\n'), 0o600); werr != nil {
				_, _ = fmt.Fprintln(os.Stderr, "agent-eval: write report: "+werr.Error())
			}
		}
	}
	if !report.OK {
		return 1
	}
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
