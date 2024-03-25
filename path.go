package main

import "path"

var (
	basePath = "/"
)

func getPathFromBase(p string) string {
	if basePath == "/" {
		return p
	}
	return p[len(basePath)-1:]
}

func buildPathFromBase(p string) string {
	if basePath == "/" {
		return p
	}
	return path.Join(basePath, p) + "/"
}
