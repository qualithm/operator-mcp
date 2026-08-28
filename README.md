# Operator MCP

[![CI](https://github.com/qualithm/operator-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/qualithm/operator-mcp/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/qualithm/operator-mcp/graph/badge.svg)](https://codecov.io/gh/qualithm/operator-mcp)

Go MCP server for agent-native provisioning of the Qualithm platform. It exposes the platform
management API — teams, members, spaces, devices, credentials, enrollments, authorities, API tokens,
automations, dashboards, observability, and read-only workspace and billing state — as MCP tools over
stdio, backed by the same `operator` client that powers the `qualithm` CLI, so the human and agent
surfaces never diverge.

## Features

- **Provisioning tools** — one MCP tool per verb over every resource: authorities, enrollments,
  credentials, devices, and API tokens.
- **Uniform result envelope** — every tool returns the same
  `{ ok, code, message, dryRun, action, data }` shape, so agents branch on one structure regardless
  of tool.
- **Per-call dry-run** — mutating tools accept `dryRun`; when set, the change is planned and the
  intended request is reported without being sent.
- **Stable error codes** — failures carry a code mirroring the CLI's exit-code contract (`auth`,
  `not_found`, `conflict`, `rate_limited`, `api`, `error`).
- **Bearer auth** — authenticates with a member API token (prefix `qmt_`).

## Installation

```bash
go install github.com/qualithm/operator-mcp/cmd/qualithm-mcp@latest
```

A container image is published to GHCR on each release:

```bash
docker pull ghcr.io/qualithm/operator-mcp:latest
```

## Quick Start

The server authenticates with a member API token (prefix `qmt_`). Provide it via `--token` or the
`QUALITHM_API_TOKEN` environment variable; point at an environment with `--url` or
`QUALITHM_API_URL` (defaults to `https://api.qualithm.com`). It speaks the Model Context Protocol
over stdio.

```bash
export QUALITHM_API_TOKEN=qmt_...
qualithm-mcp
```

Register it with an MCP-capable agent, for example:

```json
{
  "mcpServers": {
    "qualithm-operator": {
      "command": "qualithm-mcp",
      "env": {
        "QUALITHM_API_TOKEN": "qmt_..."
      }
    }
  }
}
```

| Flag      | Env                  | Description                |
| --------- | -------------------- | -------------------------- |
| `--url`   | `QUALITHM_API_URL`   | management API base URL    |
| `--token` | `QUALITHM_API_TOKEN` | member API token (`qmt_…`) |
| `--help`  | —                    | print usage and exit       |

## Tools

| Resource | Tools |
| --- | --- |
| authorities | `list_authorities` · `create_authority` · `revoke_authority` |
| enrollments | `list_enrollments` · `create_enrollment` · `revoke_enrollment` |
| credentials | `list_credentials` · `mint_credential` · `issue_cert` · `rotate_credential` · `revoke_credential` |
| devices | `list_devices` · `list_space_devices` · `get_device` · `create_device` · `update_device` · `delete_device` |
| device commands | `list_device_commands` · `send_device_command` · `get_device_capabilities` · `park_device` · `unpark_device` |
| spaces | `list_spaces` · `get_space` · `create_space` · `update_space` · `delete_space` |
| teams | `list_teams` · `create_team` · `get_team` · `update_team` · `delete_team` · `get_team_device_state` |
| members & invites | `list_team_members` · `add_team_member` · `get_team_member` · `update_team_member` · `remove_team_member` · `list_team_invites` · `create_team_invite` · `revoke_team_invite` |
| automations | `list_automations` · `create_automation` · `get_automation` · `update_automation` · `delete_automation` · `enable_automation` · `disable_automation` · `run_automation` · `list_automation_runs` · `create_automation_from_template` · `list_automation_templates` · `trigger_automation` · `create_automation_trigger_secret` |
| dashboards | `list_dashboards` · `create_dashboard` · `get_dashboard` · `update_dashboard` · `delete_dashboard` |
| observability | `get_telemetry` · `stream_events` · `get_usage` · `get_audit_log` |
| workspace & account | `get_workspace` · `get_account` · `list_capabilities` · `list_roles` · `list_sessions` · `get_session` · `get_communication_preferences` · `list_zone_spaces` |
| api tokens | `list_api_tokens` · `create_api_token` · `revoke_api_token` |
| billing (read-only) | `get_billing_summary` · `list_invoices` · `preview_tier_change` |

Money-moving billing routes (tier changes, add-ons, checkout and portal sessions) and account/session
mutations stay human-only by decision (qualithm/discussions#432). The full route-to-tool mapping,
including every excluded route's rationale, is `cmd/coverage-check/coverage.json` — CI fails when a new
platform route ships without a tool or a recorded rationale.

### Result envelope

Every tool returns the same structured payload:

| Field     | Meaning                                                    |
| --------- | ---------------------------------------------------------- |
| `ok`      | whether the call succeeded (a dry-run counts as success)   |
| `code`    | failure classification when `ok` is false                  |
| `message` | human-readable error message when `ok` is false            |
| `dryRun`  | true when a mutation was planned but not applied           |
| `action`  | the planned request (`method`, `path`) for dry-run results |
| `data`    | the resource payload returned by the API on success        |

### Error codes

| Code           | Meaning                |
| -------------- | ---------------------- |
| `auth`         | 401 / 403              |
| `not_found`    | 404                    |
| `conflict`     | 409                    |
| `rate_limited` | 429                    |
| `api`          | other non-2xx          |
| `error`        | transport / unexpected |

## Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+

### Setup

```bash
make install-tools
```

This installs local development tooling, including `golangci-lint`, `goimports`, and `govulncheck`.

> **Note:** Tools are installed to `$GOPATH/bin` (typically `~/go/bin`). Make sure that directory is
> on your `$PATH`, otherwise the installed binaries won't be found.

### Building

```bash
make build
```

### Testing

```bash
make test              # unit tests with race detector
make test-coverage     # with coverage report
```

### Agent eval

`cmd/agent-eval` is the end-to-end agent-native eval: it drives the documented agent path against a
live environment — create an enrollment via the MCP tools, claim the device, connect to the gateway,
publish telemetry, and read it back — and records every step where the agent had to supply information
the surfaces did not provide (the friction log). The `Agent Eval` workflow runs it weekly against
sandbox and files a deduped issue on failure.

```bash
go build -o /tmp/qualithm-mcp ./cmd/qualithm-mcp
go run ./cmd/agent-eval -server-bin /tmp/qualithm-mcp -token qmt_...
```

### Linting & Formatting

```bash
make lint
make fmt
make vet
```

### Security Tooling

```bash
make audit   # govulncheck
make gosec   # standalone gosec scan
make lint    # golangci-lint (includes gosec checks via .golangci.yaml)
```

Daily CI security audit runs both tools in `.github/workflows/audit.yaml`.

## Minimum Supported Go Version

Go 1.26+.

## License

Apache-2.0
