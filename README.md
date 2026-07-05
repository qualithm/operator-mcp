# Operator MCP

<!-- TODO: uncomment badges after first publish
[![CI](https://github.com/qualithm/operator-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/qualithm/operator-mcp/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/qualithm/operator-mcp/graph/badge.svg)](https://codecov.io/gh/qualithm/operator-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/qualithm/operator-mcp)](https://goreportcard.com/report/github.com/qualithm/operator-mcp)
-->

Go MCP server for agent-native provisioning of the Qualithm platform. It exposes the platform
management API — authorities, enrollments, credentials, devices, and API tokens — as MCP tools over
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

| Resource    | Tools                                                                                                      |
| ----------- | ---------------------------------------------------------------------------------------------------------- |
| authorities | `list_authorities` · `create_authority` · `revoke_authority`                                               |
| enrollments | `list_enrollments` · `create_enrollment` · `revoke_enrollment`                                             |
| credentials | `list_credentials` · `mint_credential` · `issue_cert` · `rotate_credential` · `revoke_credential`          |
| devices     | `list_devices` · `list_space_devices` · `get_device` · `create_device` · `update_device` · `delete_device` |
| api tokens  | `list_api_tokens` · `create_api_token` · `revoke_api_token`                                                |

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

## Publishing

Tagged releases are automatically built and published to GHCR when CI passes on `main`. The binary
is consumed directly from the Git tag (`vX.Y.Z`) via `go install`; no separate registry publish step
is required.

## Minimum Supported Go Version

Go 1.26+.

## License

Apache-2.0

## CI & Branch Protection

The `.github/workflows/ci.yaml` workflow and the `main` / `test` branch rulesets are generated by
[dx](https://github.com/qualithm/dx). To change CI for this repo, edit the relevant archetype in
`dx/ci-templates/` and run `dx ci sync`; do not edit `ci.yaml` directly. The umbrella job at the end
of the workflow supplies the single required status check (`CI Required`).
