package main

import (
	"log"
	"net/http"
	"strings"
)

type webHandler struct {
}

// ServeHTTP 用于HTTP服务
func (p *webHandler) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) < len(basePath) {
		rsp.Header()
		http.Redirect(rsp, req, buildPathFromBase("/"), http.StatusFound)
		return
	}

	absPath := getPathFromBase(req.URL.Path)
	switch absPath {
	case "/health":
		handleHealth(rsp, req)
		return
	case "/":
		handleProfileHome(rsp, req)
		return
	case "/favicon.ico":
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	if strings.HasPrefix(absPath, "/proxy/") {
		handleProxy(rsp, req)
		return
	}

	pathParts := strings.Split(absPath, "/")
	var profileId string
	if len(pathParts) >= 2 {
		profileId = pathParts[1]
	}
	if len(profileId) <= 0 {
		rsp.Header()
		http.Redirect(rsp, req, buildPathFromBase("/"), http.StatusFound)
		return
	}

	var pathHandleMap any
	var ok bool
	if pathHandleMap, ok = profileIdPathHandleMap.Load(profileId); !ok {
		if !tryLoadProfile(profileId) {
			rsp.WriteHeader(http.StatusNotFound)
			return
		}
		if pathHandleMap, ok = profileIdPathHandleMap.Load(profileId); !ok {
			log.Println("handle still missing after load profile ", profileId)
			rsp.WriteHeader(http.StatusNotFound)
			return
		}
	}
	handleProfile(rsp, req, profileId, pathHandleMap.(map[string]http.Handler))
	return
}
