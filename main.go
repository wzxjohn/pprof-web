package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

var (
	tmpPath       = "/tmp/pprof-web"
	listenAddress = ":8080"
	logLevel      = "info"
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

	v := os.Getenv("PPROF_LOG_LEVEL")
	if len(v) > 0 {
		logLevel = v
	}

	flag.StringVar(&tmpPath, "t", tmpPath, "")
	flag.StringVar(&listenAddress, "l", listenAddress, "")
	flag.StringVar(&basePath, "b", basePath, "")
	flag.StringVar(&logLevel, "v", logLevel, "")
	flag.Parse()

	initLogger(logLevel)

	if tmpPath[len(tmpPath)-1] != '/' {
		tmpPath += "/"
	}
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		err = os.MkdirAll(tmpPath, 0755)
		if err != nil {
			slog.Error("failed to create tmp directory", "path", tmpPath, "error", err)
			os.Exit(1)
		}
	} else if err != nil {
		slog.Error("failed to stat tmp directory", "path", tmpPath, "error", err)
		os.Exit(1)
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

	slog.Info("starting http server", "address", listenAddress, "tmpDir", tmpPath, "basePath", basePath, "logLevel", logLevel)

	err := http.ListenAndServe(listenAddress, &webHandler{})
	if err != nil {
		slog.Error("http server exited", "error", err)
		os.Exit(1)
	}
}

func initLogger(level string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})
	slog.SetDefault(slog.New(handler))
	// Also redirect stdlib log output to slog-compatible format
	log.SetFlags(0)
}
