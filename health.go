package main

import "net/http"

func handleHealth(rsp http.ResponseWriter, _ *http.Request) {
	rsp.Write([]byte("ok"))
}
