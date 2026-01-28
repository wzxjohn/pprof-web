# PProf-Web

[![Docker Image CI](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml/badge.svg)](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml)

Web proxy for Go pprof endpoints - access pprof debug endpoints through firewalls with an interactive UI.

PProf-Web can be deployed in a restricted network zone and expose a single web endpoint to proxy pprof requests, providing the full pprof web interface for profiling remote Go applications.

## Quick Start

### Docker

```bash
docker run -p 8080:8080 ghcr.io/wzxjohn/pprof-web
```

### Build from Source

```bash
go build
./pprof-web
```

Open http://localhost:8080, enter the target IP, port, and profile type, then view the interactive pprof UI.

## Features

- **Profile Types**: CPU, heap, and goroutine profiles
- **Interactive UI**: Full pprof web interface with flame graphs and call graphs
- **Proxy Mode**: Direct proxy to `/debug/pprof/*` endpoints at `/proxy/{ip}/{port}/debug/pprof/`
- **Persistent Storage**: Profiles saved locally for later viewing
- **Reverse Proxy Support**: Configurable base path for deployment behind nginx/ingress

## Configuration

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-l` | `PPROF_LISTEN_ADDRESS` | `:8080` | Listen address |
| `-t` | `PPROF_TMP_PATH` | `/tmp/pprof-web` | Temp directory for profile storage |
| `-b` | `PPROF_BASE_PATH` | `/` | Base path for reverse proxy deployments |

### Example

```bash
./pprof-web -l :9090 -t /data/profiles -b /pprof/
```

## Deployment

### Kubernetes

Helm chart available in `manifests/charts/pprof-web/`.

### Nginx Reverse Proxy

When using nginx as a reverse proxy, set `proxy_read_timeout` larger than the max profile duration (60s) since the server cannot send data while profiling.

```nginx
location /cluster-1/ {
    rewrite ^/cluster-1(/.*)$ $1 break;
    proxy_redirect / /cluster-1/;
    proxy_read_timeout 120s;
    proxy_pass http://pprof-web:8080/;
}
```

Start pprof-web with matching base path:

```bash
./pprof-web -b /cluster-1/
```

## Design Goals

- Use only the official `github.com/google/pprof` tool as dependency
- Minimal implementation of pprof interfaces

## Roadmap

- [ ] Auto-delete inactive profiles
- [ ] Improved logging
- [ ] Unit tests
- [ ] Enhanced web UI

## Stargazers over time

[![Stargazers over time](https://starchart.cc/wzxjohn/pprof-web.svg)](https://starchart.cc/wzxjohn/pprof-web)
