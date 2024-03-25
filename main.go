package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var (
	tmpPath       = "/tmp/pprof-web"
	listenAddress = ":8080"
)

func main() {
	t := os.Getenv("PPROF_TMP_PATH")
	if len(t) > 0 {
		tmpPath = t
	}

	l := os.Getenv("PPROF_LISTEN_ADDRESS")
	if len(l) > 0 {
		listenAddress = l
	}

	b := os.Getenv("PPROF_BASE_PATH")
	if len(b) > 0 {
		basePath = b
	}

	flag.StringVar(&tmpPath, "t", tmpPath, "")
	flag.StringVar(&listenAddress, "l", listenAddress, "")
	flag.StringVar(&basePath, "b", basePath, "")
	flag.Parse()

	if tmpPath[len(tmpPath)-1] != '/' {
		tmpPath += "/"
	}
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		err = os.MkdirAll(tmpPath, 0755)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	if len(basePath) == 0 {
		basePath = "/"
	}
	if basePath[0] != '/' {
		basePath = "/" + basePath
	}
	if basePath[len(basePath)-1] != '/' {
		basePath += "/"
	}

	log.Printf("start http server at %s using tmp dir %s base path %s", listenAddress, tmpPath, basePath)

	err := http.ListenAndServe(listenAddress, &webHandler{})
	if err != nil {
		panic(err)
	}
}
