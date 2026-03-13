# PProf-Web

[![Docker Image CI](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml/badge.svg)](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml)

A web proxy for Go pprof endpoints. Deploy PProf-Web in a restricted network zone to expose a single web endpoint that proxies pprof requests, giving you the full interactive pprof UI for profiling remote Go applications — without opening additional firewall rules.

## Quick Start

### Docker (recommended)

```bash
docker run -p 8080:8080 ghcr.io/wzxjohn/pprof-web
```

Also available on Docker Hub:

```bash
docker run -p 8080:8080 wzxjohn/pprof-web
```

### Build from Source

**Requirements:** Go 1.26+, [Graphviz](https://graphviz.org/) (for graph rendering)

```bash
go build
./pprof-web
```

Open http://localhost:8080, enter the target IP, port, and profile type, then explore the interactive pprof UI.

## Features

- **Profile Types** — CPU (with configurable duration up to 60s), heap, and goroutine profiles
- **Interactive UI** — Full pprof web interface with flame graphs, call graphs, and source views
- **Proxy Mode** — Transparent proxy to remote `/debug/pprof/*` endpoints at `/proxy/{ip}/{port}/debug/pprof/`
- **Persistent Storage** — Profiles saved locally for later viewing
- **Reverse Proxy Support** — Configurable base path for deployment behind nginx or ingress controllers
- **Health Check** — `/health` endpoint for liveness/readiness probes
- **Minimal Dependencies** — Only the official `github.com/google/pprof` library, compiled as a static binary with CGO disabled

## How It Works

```
┌────────┐         ┌───────────┐         ┌─────────────────┐
│ Browser │ ──1──▶ │ PProf-Web │ ──2──▶  │ Go App          │
│         │ ◀──4── │           │ ◀──3──  │ /debug/pprof/*  │
└────────┘         └───────────┘         └─────────────────┘
```

1. User submits the target's IP, port, and profile type via the web form
2. PProf-Web fetches the profile from `http://{ip}:{port}/debug/pprof/{type}`
3. The profile data is returned and saved to the temp directory
4. User is redirected to `/{profileId}/` for the interactive pprof UI

## Configuration

| Flag | Environment Variable    | Default          | Description                          |
|------|-------------------------|------------------|--------------------------------------|
| `-l` | `PPROF_LISTEN_ADDRESS`  | `:8080`          | Listen address                       |
| `-t` | `PPROF_TMP_PATH`        | `/tmp/pprof-web` | Directory for storing fetched profiles |
| `-b` | `PPROF_BASE_PATH`       | `/`              | Base path for reverse proxy deployments |

Flags take precedence over environment variables.

### Example

```bash
./pprof-web -l :9090 -t /data/profiles -b /pprof/
```

## Deployment

### Kubernetes (Helm)

A Helm chart is available in `manifests/charts/pprof-web/`:

```bash
helm install pprof-web manifests/charts/pprof-web/
```

The chart supports:
- Configurable replicas, resources, and service type
- Optional ingress with TLS
- Optional PersistentVolume/PersistentVolumeClaim for profile storage
- Readiness probe on `/health` and liveness probe via TCP socket

See [`manifests/charts/pprof-web/values.yaml`](manifests/charts/pprof-web/values.yaml) for all available options.

### Nginx Reverse Proxy

When deploying behind nginx, set `proxy_read_timeout` higher than the max profile duration (60s) since the server blocks while collecting a CPU profile:

```nginx
location /cluster-1/ {
    rewrite ^/cluster-1(/.*)$ $1 break;
    proxy_redirect / /cluster-1/;
    proxy_read_timeout 120s;
    proxy_pass http://pprof-web:8080/;
}
```

Start PProf-Web with the matching base path:

```bash
./pprof-web -b /cluster-1/
```

## Security

The proxy endpoint (`/proxy/`) only allows requests to the standard Go pprof paths. All other paths are rejected. This prevents the proxy from being used as an open HTTP relay.

## Design Goals

- Use only the official `github.com/google/pprof` tool as dependency
- Minimal implementation of pprof interfaces

## Roadmap

- [ ] Auto-delete inactive profiles
- [X] Improved logging
- [ ] Unit tests
- [X] Enhanced web UI

## Stargazers over time

[![Stargazers over time](https://starchart.cc/wzxjohn/pprof-web.svg)](https://starchart.cc/wzxjohn/pprof-web)
