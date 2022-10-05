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
	flag.StringVar(&tmpPath, "t", tmpPath, "")
	flag.StringVar(&listenAddress, "l", listenAddress, "")
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

	log.Printf("start http server at %s using tmp dir %s", listenAddress, tmpPath)

	err := http.ListenAndServe(listenAddress, &profileProxy{})
	if err != nil {
		panic(err)
	}
}
