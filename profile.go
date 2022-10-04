package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/pprof/driver"
	"github.com/google/uuid"
)

const (
	ProfileTypeCPU    = "CPU"
	ProfileTypeMemory = "MEM"
)

var (
	profileIdPathHandleMap sync.Map
	idProfileIdMap         sync.Map
)

type profileProxy struct {
}

func (p *profileProxy) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		handleHome(rsp, req)
		return
	}

	if req.URL.Path == "/favicon.ico" {
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	pathParts := strings.Split(req.URL.Path, "/")
	var profileId string
	if len(pathParts) >= 2 {
		profileId = pathParts[1]
	}
	if len(profileId) <= 0 {
		rsp.Header()
		http.Redirect(rsp, req, "/", http.StatusFound)
		return
	}

	var pathHandleMap any
	var ok bool
	if pathHandleMap, ok = profileIdPathHandleMap.Load(profileId); !ok {
		if !tryLoadProfile(profileId) {
			return
		}
	}
	handleProfile(rsp, req, profileId, pathHandleMap.(map[string]http.Handler))
	return

}

func handleHome(rsp http.ResponseWriter, req *http.Request) {
	ip := req.URL.Query().Get("ip")
	portStr := req.URL.Query().Get("port")
	secondsStr := req.URL.Query().Get("seconds")
	profileType := req.URL.Query().Get("type")
	if len(ip) > 0 && len(portStr) > 0 {
		var port int
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return
		}

		var seconds int
		if secondsStr == "" {
			seconds = 30
		} else {
			seconds, err = strconv.Atoi(secondsStr)
			if err != nil {
				return
			}
		}
		if profileType == "" {
			profileType = ProfileTypeCPU
		}
		profileId := newProfileId(ip, port)
		err = newProfile(profileId, ip, port, seconds, profileType)
		if err != nil {
			rsp.Write([]byte("fetch failed.\n" + err.Error()))
			return
		}
		http.Redirect(rsp, req, "./"+profileId+"/", http.StatusFound)
		return
	}

	_, _ = rsp.Write([]byte(`<html>
<head>
<title>PProf Proxy</title>
</head>
<body>
<form action="./">
<label for="ip">IP:</label><br>
<input type="text" id="ip" name="ip" size="20" placeholder="10.0.0.1" autofocus required><br>
<label for="port">Port:</label><br>
<input type="number" id="port" name="port" size="20" placeholder="8000" min="1" max="65535" required><br>
<label for="seconds">Seconds:</label><br>
<input type="number" id="seconds" name="seconds" size="20" placeholder="30" min="1" max="60"><br>
<br>
<input type="submit" value="Profile!" onClick="this.form.submit(); this.disabled=true; this.value='Profiling…';">
</form>
</body>
</html>`))
	return
}

func handleProfile(rsp http.ResponseWriter, req *http.Request, profileId string, pathHandle map[string]http.Handler) {
	realPath := req.URL.Path[len(profileId)+1:]
	var handle http.Handler
	var ok bool
	if handle, ok = pathHandle[realPath]; !ok {
		handle = pathHandle["/"]
	}
	handle.ServeHTTP(rsp, req)
}

func newProfile(profileId, ip string, port, seconds int, profileType string) error {
	_id := nextId()
	log.Println("profile ", profileId, "assigned id ", _id)
	idProfileIdMap.Store(_id, profileId)

	profilePath, err := fetchProfile(profileId, ip, port, seconds, profileType)
	if err != nil {
		return err
	}

	o := newOption(_id, ip, port, profilePath)
	err = driver.PProf(o)
	if err != nil {
		return err
	}
	return nil
}

func tryLoadProfile(profileId string) bool {
	ip, port, realId := parseProfileId(profileId)
	if ip == "" || port == 0 || realId == "" {
		log.Println("wrong profile id: ", profileId)
		return false
	}

	_id := nextId()
	log.Println("profile ", profileId, "assigned id ", _id)
	idProfileIdMap.Store(_id, profileId)

	profilePath := getProfilePath(profileId)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		// profile not exist
		return false
	}

	o := newOption(_id, ip, port, profilePath)
	err := driver.PProf(o)
	if err != nil {
		return false
	}
	return true
}

func newOption(id int, ip string, port int, profilePath string) *driver.Options {
	return &driver.Options{
		Flagset: &webFlagSet{
			strings: map[string]string{"http": "0.0.0.0:" + strconv.Itoa(id)},
			args:    []string{ip + ":" + strconv.Itoa(port), profilePath},
		},
		Obj:        &webObjTool{},
		UI:         &webUI{},
		HTTPServer: pprofHTTPServer,
	}
}

func fetchProfile(profileId, ip string, port, seconds int, profileType string) (string, error) {
	url := fmt.Sprintf("http://%s:%d/debug/pprof/profile?seconds=%d", ip, port, seconds)
	switch profileType {
	case ProfileTypeMemory:
		url += ""
		break
	}

	client := &http.Client{
		Timeout: time.Duration(seconds)*time.Second + 5*time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Println("http fetch: ", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		log.Println("http fetch get status: ", resp.StatusCode)
		return "", fmt.Errorf("wrong http status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	profilePath := getProfilePath(profileId)
	f, err := os.Create(profilePath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	log.Println("save profile in ", profilePath)
	return profilePath, nil
}

func getProfilePath(profileId string) string {
	return tmpPath + profileId + ".pprof"
}

func pprofHTTPServer(args *driver.HTTPServerArgs) error {
	log.Println("start http server for ", args.Port)
	profileId, ok := idProfileIdMap.Load(args.Port)
	if ok {
		log.Println("match id ", args.Port, "to profile id ", profileId)
		profileIdPathHandleMap.Store(profileId.(string), args.Handlers)
	}
	return nil
}

func newProfileId(ip string, port int) string {
	u := uuid.NewString()
	profileId := ip + "_" + strconv.Itoa(port) + "_" + u
	return profileId
}

func parseProfileId(profileId string) (ip string, port int, id string) {
	idParts := strings.Split(profileId, "_")

	if len(idParts) >= 3 {
		var err error
		port, err = strconv.Atoi(idParts[1])
		if err != nil {
			log.Println("wrong profile id format: ", profileId)
			return
		}
		ip = idParts[0]
		id = idParts[2]
	}
	return
}
