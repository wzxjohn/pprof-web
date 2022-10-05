# PProf-Web

[![Docker Image CI](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml/badge.svg)](https://github.com/wzxjohn/pprof-web/actions/workflows/docker-image.yml)

---
PProd web is a web proxy for pprof endpoints.
Sometimes we have firewalls between net areas and cannot retch the pprof debug endpoint directly.
PProf-Web can be deployed in this area and expose only one web endpoint to proxy the pprof request.

# Goals

---

- ONLY use official pprof tool as go mod dependency
- Implement interfaces in official pprof tool as little as possible


# Feature

---

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

# Stargazers over time

---

[![Stargazers over time](https://starchart.cc/wzxjohn/pprof-web.svg)](https://starchart.cc/wzxjohn/pprof-web)
