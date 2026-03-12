# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pprof-web is a web proxy for Go pprof endpoints. It allows accessing pprof debug endpoints through firewalls by deploying this service in a network zone and exposing a single web endpoint.

## Build and Run Commands

```bash
# Build
go build

# Run (default: listens on :8080, uses /tmp/pprof-web for temp storage)
./pprof-web

# Run with flags
./pprof-web -l :8080 -t /tmp/pprof-web -b /base-path/

# Run tests (no test files exist yet)
go test ./...

# Docker
docker build -t pprof-web .
docker run -p 8080:8080 pprof-web
```

## Configuration

Environment variables (can also be set via flags):
- `PPROF_LISTEN_ADDRESS` / `-l`: Listen address (default: `:8080`)
- `PPROF_TMP_PATH` / `-t`: Temp directory for profile storage (default: `/tmp/pprof-web`)
- `PPROF_BASE_PATH` / `-b`: Base path for reverse proxy deployments (default: `/`)

## Architecture

The application is a single-binary HTTP server that wraps the official `github.com/google/pprof` tool. Built with CGO disabled for static binary portability. Graphviz is required at runtime for graph generation (installed in Docker image).

### Core Components

- **handler.go**: Main HTTP router (`webHandler`). Routes requests to `/health`, `/`, `/proxy/*`, or profile views based on URL path.

- **profile.go**: Profile fetching and management. Fetches profiles from remote pprof endpoints, stores them in temp directory, and initializes pprof's web UI handlers. Uses `sync.Map` for thread-safe storage of profile handlers.

- **proxy.go**: Transparent proxy for pprof endpoints at `/proxy/{ip}/{port}/debug/pprof/*`. Whitelists only standard pprof endpoints for security.

- **webui.go**, **flag.go**, **symbolizer.go**: Minimal implementations of pprof driver interfaces (`driver.UI`, `driver.FlagSet`, `driver.Symbolize`) to integrate pprof as a library rather than CLI. The key pattern: `driver.PProf()` is called with these custom implementations plus a `pprofHTTPServer` callback that captures registered HTTP handlers into a `sync.Map` instead of starting a standalone server.

- **path.go**: Base path utilities (`getPathFromBase`, `buildPathFromBase`) enabling deployment behind reverse proxies with URL rewriting.

### Request Flow

1. User submits IP/port/type via web form at `/`
2. Server fetches profile from `http://{ip}:{port}/debug/pprof/{type}`
3. Profile saved to temp directory, pprof handlers registered via `sync.Map`
4. User redirected to `/{profileId}/` for interactive pprof UI

### Profile Types

Supports `cpu`, `heap`, and `goroutine` profile types. CPU profiles accept a `seconds` parameter (max 60).

## Deployment

Helm chart available at `manifests/charts/pprof-web/` for Kubernetes deployment. CI/CD via GitHub Actions builds and pushes Docker images to both Docker Hub and GHCR on pushes to `master` or tags.
