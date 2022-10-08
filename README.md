# PProf-Web

[![Docker Image CI](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml/badge.svg)](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml)

PProd web is a web proxy for pprof endpoints.
Sometimes we have firewalls between net areas and cannot retch the pprof debug endpoint directly.
PProf-Web can be deployed in this area and expose only one web endpoint to proxy the pprof request.

# Goals

- ONLY use official pprof tool as go mod dependency
- Implement interfaces in official pprof tool as little as possible

# Feature

- [x] Fetch remote CPU profile
- [x] Load profile from local tmp storage
- [ ] Auto delete inactive profile
- [ ] Better log
- [x] Dockerfile
- [x] Helm chart
- [ ] Unit test
- [ ] Support memory profile
- [ ] Improve web UI
- [x] Proxy for all pprof endpoint

# Usage

## Nginx proxy as sub folder

Be careful of the `proxy_read_timeout` option.
Because server cannot send any response data during profile,
this option must larger than max profile seconds (usually 60s).

```nginx
location /cluster-1/ {
    rewrite ^/cluster-1(/.*)$ $1 break;
    proxy_redirect / /cluster-1/;
    proxy_read_timeout 120s;
    proxy_pass http://1.1.1.1:8080/;
}

location /cluster-2/ {
    rewrite ^/cluster-2(/.*)$ $1 break;
    proxy_redirect / /cluster-2/;
    proxy_read_timeout 120s;
    proxy_pass http://2.2.2.2:8080/;
}
```

# Stargazers over time

[![Stargazers over time](https://starchart.cc/wzxjohn/pprof-web.svg)](https://starchart.cc/wzxjohn/pprof-web)
