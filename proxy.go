package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	allowedEndpoint = map[string]struct{}{
		"/debug/pprof/":             {},
		"/debug/pprof/allocs":       {},
		"/debug/pprof/block":        {},
		"/debug/pprof/cmdline":      {},
		"/debug/pprof/goroutine":    {},
		"/debug/pprof/heap":         {},
		"/debug/pprof/mutex":        {},
		"/debug/pprof/profile":      {},
		"/debug/pprof/threadcreate": {},
		"/debug/pprof/trace":        {},
	}
	bufferPool = sync.Pool{New: func() any {
		return make([]byte, 32*1024)
	}}
)

// handleProxy request like /proxy/1.2.3.4/8000/debug/pprof
func handleProxy(rsp http.ResponseWriter, req *http.Request) {
	absPath := getPathFromBase(req.URL.Path)
	pathParts := strings.Split(absPath, "/")
	var ipStr, portStr string
	if len(pathParts) < 4 {
		log.Println("missing proxy endpoint: ", absPath)
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}
	ipStr = pathParts[2]
	portStr = pathParts[3]
	endpoint := absPath[8+len(ipStr)+len(portStr):]
	if endpoint == "" || endpoint == "/" {
		req.URL.Path += "/debug/pprof/"
		http.Redirect(rsp, req, req.URL.String(), http.StatusFound)
		return
	}
	if endpoint == "/debug/pprof" {
		req.URL.Path += "/"
		http.Redirect(rsp, req, req.URL.String(), http.StatusFound)
		return
	}
	if _, ok := allowedEndpoint[endpoint]; !ok {
		log.Println("endpoint not allowed: ", endpoint)
		rsp.WriteHeader(http.StatusForbidden)
		return
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Println("invalid target ip: ", ipStr)
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println("invalid target port: ", portStr)
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}

	timeout := 65
	secondsStr := req.URL.Query().Get(secondsQueryParam)
	if secondsStr != "" {
		timeout, err = strconv.Atoi(secondsStr)
		if err != nil {
			log.Println("invalid seconds param: ", secondsStr)
			rsp.WriteHeader(http.StatusBadRequest)
			return
		}
		if timeout > 60 {
			timeout = 65
			req.URL.Query().Set(secondsQueryParam, "60")
		} else {
			timeout += 5
		}
	}

	doProxy(ipStr, portStr, endpoint, time.Duration(timeout)*time.Second, rsp, req)
}

func doProxy(ip, port, endpoint string, timeout time.Duration, rsp http.ResponseWriter, req *http.Request) {
	client := &http.Client{
		Timeout: timeout,
	}
	targetURL := fmt.Sprintf("http://%s:%s%s?%s", ip, port, endpoint, req.URL.RawQuery)
	proxyReq, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		log.Println("error while new http request: ", err)
		rsp.WriteHeader(http.StatusInternalServerError)
		return
	}

	proxyReq.Header = req.Header.Clone()
	proxyRsp, err := client.Do(proxyReq)
	if err != nil {
		log.Println("error while fetch target: ", err)
		rsp.WriteHeader(http.StatusBadGateway)
		return
	}
	defer proxyRsp.Body.Close()

	for k, vs := range proxyRsp.Header {
		for _, v := range vs {
			rsp.Header().Set(k, v)
		}
	}
	rsp.WriteHeader(proxyRsp.StatusCode)

	_, err = copyResponse(rsp, proxyRsp.Body)
	if err != nil {
		log.Println("write response failed: ", err)
		return
	}
}

func copyResponse(dst io.Writer, src io.Reader) (int64, error) {
	bufI := bufferPool.Get()
	defer bufferPool.Put(bufI)
	buf := bufI.([]byte)
	var written int64

	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			log.Println("ReverseProxy read error during body copy: ", rerr)
		}
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return written, werr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			return written, rerr
		}
	}
}
