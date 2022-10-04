package main

import (
	"fmt"
	"regexp"

	"github.com/google/pprof/driver"
)

type webObjTool struct {
}

func (w *webObjTool) Open(_ string, _, limit, _ uint64, _ string) (driver.ObjFile, error) {
	// Only return file obj in first open request to display correct file name
	// github.com/google/pprof@v0.0.0-20220829040838-70bd9ae97f40/internal/driver/cli.go:101
	// first open always use ^uint64(0) as limit
	if limit == ^uint64(0) {
		return &webObjFile{}, nil
	}
	return nil, fmt.Errorf("file not found")
}

func (w *webObjTool) Disasm(_ string, _, _ uint64, _ bool) ([]driver.Inst, error) {
	return nil, fmt.Errorf("disassembly not supported")
}

type webObjFile struct {
}

func (w *webObjFile) Name() string {
	return ""
}

func (w *webObjFile) ObjAddr(uint64) (uint64, error) {
	return 0, fmt.Errorf("obj addr not supported")
}

func (w *webObjFile) BuildID() string {
	return ""
}

func (w *webObjFile) SourceLine(uint64) ([]driver.Frame, error) {
	return nil, fmt.Errorf("source line not supported")
}

func (w *webObjFile) Symbols(*regexp.Regexp, uint64) ([]*driver.Sym, error) {
	return nil, fmt.Errorf("symbols not supported")
}

func (w *webObjFile) Close() error {
	return nil
}
