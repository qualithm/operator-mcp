// Command coverage-check fails when the platform management API's routes drift
// from the MCP tool surface: a route ships without a tool counterpart, without a
// recorded no-tool rationale, or without a gap-issue reference in
// internal/server/coverage.json. Run from the repo root; -platform points at a
// platform checkout (CI clones qualithm/platform alongside).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	platformDir = flag.String("platform", "../platform", "path to a qualithm/platform checkout")
	ledgerPath  = flag.String("ledger", "internal/server/coverage.json", "path to the coverage ledger")
)

// ledgerEntry classifies one platform management route. Exactly one of Tool,
// NoTool, Issue must be set: Tool when an MCP tool wraps the route, NoTool with
// the exclusion category when the route is not builder-reachable, Issue with the
// tracking gap-issue number when coverage is planned but not yet implemented.
type ledgerEntry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Tool   string `json:"tool,omitempty"`
	NoTool string `json:"noTool,omitempty"`
	Issue  int    `json:"issue,omitempty"`
}

type ledger struct {
	Routes []ledgerEntry `json:"routes"`
}

var (
	// http.post( with an optional run of // comment lines, then the path
	// literal on its own line — the shape every platform route file uses.
	routeRe = regexp.MustCompile(`(?s)http\.(get|post|patch|delete)\(\s*(?://[^\n]*\s*)*"(/[^"]*)"`)
	// mcp.AddTool(srv, &mcp.Tool{ Name: "x", ... — registered tools only, so
	// server names in tests and literals elsewhere never count.
	toolRe = regexp.MustCompile(`(?s)AddTool\(srv,\s*&mcp\.Tool\{\s*Name:\s*"([a-z_]+)"`)
)

func main() {
	flag.Parse()
	if err := run(*platformDir, *ledgerPath, "internal/server"); err != nil {
		fmt.Fprintf(os.Stderr, "coverage-check: %v\n", err)
		os.Exit(1)
	}
}

func run(platform, ledgerFile, serverDir string) error {
	scanned, err := scanRoutes(platform)
	if err != nil {
		return err
	}
	tools, err := scanTools(serverDir)
	if err != nil {
		return err
	}
	lg, err := loadLedger(ledgerFile)
	if err != nil {
		return err
	}

	var problems []string

	scannedKeys := map[string]bool{}
	for _, r := range scanned {
		scannedKeys[r.Method+" "+r.Path] = true
	}
	ledgerKeys := map[string]ledgerEntry{}
	for _, e := range lg.Routes {
		key := e.Method + " " + e.Path
		if _, dup := ledgerKeys[key]; dup {
			problems = append(problems, fmt.Sprintf("duplicate ledger entry: %s", key))
		}
		ledgerKeys[key] = e
		set := 0
		for _, present := range []bool{e.Tool != "", e.NoTool != "", e.Issue != 0} {
			if present {
				set++
			}
		}
		if set != 1 {
			problems = append(problems, fmt.Sprintf("%s: set exactly one of tool, noTool, issue", key))
		}
		if !scannedKeys[key] {
			problems = append(problems, fmt.Sprintf("stale ledger entry (route not in platform): %s", key))
		}
		if e.Tool != "" && !tools[e.Tool] {
			problems = append(problems, fmt.Sprintf("ledger tool %q (%s) is not a registered MCP tool", e.Tool, key))
		}
	}

	for _, r := range scanned {
		key := r.Method + " " + r.Path
		if _, ok := ledgerKeys[key]; !ok {
			problems = append(problems, fmt.Sprintf("unclassified platform route: %s — add a tool, a noTool rationale, or a gap issue to %s", key, ledgerFile))
		}
	}

	claimed := map[string]bool{}
	for _, e := range lg.Routes {
		if e.Tool != "" {
			claimed[e.Tool] = true
		}
	}
	for name := range tools {
		if !claimed[name] {
			problems = append(problems, fmt.Sprintf("registered tool %q has no route in %s", name, ledgerFile))
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		for _, p := range problems {
			fmt.Fprintf(os.Stderr, "drift: %s\n", p)
		}
		return fmt.Errorf("%d drift problem(s)", len(problems))
	}

	covered, excluded, gaps := 0, 0, 0
	for _, e := range lg.Routes {
		switch {
		case e.Tool != "":
			covered++
		case e.NoTool != "":
			excluded++
		default:
			gaps++
		}
	}
	fmt.Printf("coverage-check: %d routes — %d covered, %d excluded, %d tracked gaps\n", len(lg.Routes), covered, excluded, gaps)
	return nil
}

func scanRoutes(platform string) ([]ledgerEntry, error) {
	dir := filepath.Join(platform, "src", "management")
	files, err := filepath.Glob(filepath.Join(dir, "*.ts"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no route files under %s (is -platform a qualithm/platform checkout?)", dir)
	}
	var routes []ledgerEntry
	for _, f := range files {
		if strings.HasSuffix(f, "index.ts") {
			continue
		}
		src, err := os.ReadFile(f) // #nosec G304 -- paths come from a repo-owned directory glob, not user input
		if err != nil {
			return nil, err
		}
		for _, m := range routeRe.FindAllStringSubmatch(string(src), -1) {
			routes = append(routes, ledgerEntry{Method: strings.ToUpper(m[1]), Path: m[2]})
		}
	}
	return routes, nil
}

func scanTools(dir string) (map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	tools := map[string]bool{}
	for _, f := range files {
		src, err := os.ReadFile(f) // #nosec G304 -- paths come from a repo-owned directory glob, not user input
		if err != nil {
			return nil, err
		}
		for _, m := range toolRe.FindAllStringSubmatch(string(src), -1) {
			tools[m[1]] = true
		}
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("no registered tools found under %s", dir)
	}
	return tools, nil
}

func loadLedger(path string) (ledger, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- the ledger path is a repo-owned flag default
	if err != nil {
		return ledger{}, err
	}
	var lg ledger
	if err := json.Unmarshal(raw, &lg); err != nil {
		return ledger{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(lg.Routes) == 0 {
		return ledger{}, fmt.Errorf("%s has no routes", path)
	}
	return lg, nil
}
