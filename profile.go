package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/pprof/driver"
	"github.com/google/uuid"
)

// ProfileType TODO
type ProfileType uint8

const (
	// ProfileTypeCPU TODO
	ProfileTypeCPU ProfileType = iota
	// ProfileTypeHeap TODO
	ProfileTypeHeap
	// ProfileTypeGoroutine TODO
	ProfileTypeGoroutine
)

var (
	profileIdPathHandleMap sync.Map
	idProfileIdMap         sync.Map
)

func handleProfileHome(rsp http.ResponseWriter, req *http.Request) {
	ip := req.URL.Query().Get("ip")
	portStr := req.URL.Query().Get("port")
	secondsStr := req.URL.Query().Get(secondsQueryParam)
	profileTypeStr := req.URL.Query().Get("type")
	if len(ip) > 0 && len(portStr) > 0 {
		netIp := net.ParseIP(ip)
		if netIp == nil {
			rsp.WriteHeader(http.StatusBadRequest)
			rsp.Write([]byte("invalid ip"))
			return
		}

		var port int
		port, err := strconv.Atoi(portStr)
		if err != nil {
			rsp.WriteHeader(http.StatusBadRequest)
			rsp.Write([]byte("invalid port"))
			return
		}

		var seconds int
		if secondsStr == "" {
			seconds = 30
		} else {
			seconds, err = strconv.Atoi(secondsStr)
			if err != nil {
				rsp.WriteHeader(http.StatusBadRequest)
				rsp.Write([]byte("invalid seconds"))
				return
			}
			if seconds > 60 {
				seconds = 60
			}
		}
		profileType := ProfileTypeCPU
		switch profileTypeStr {
		case "cpu":
			profileType = ProfileTypeCPU
		case "heap":
			profileType = ProfileTypeHeap
		case "goroutine":
			profileType = ProfileTypeGoroutine
		default:
			rsp.WriteHeader(http.StatusBadRequest)
			rsp.Write([]byte("wrong profile type"))
		}
		profileId := newProfileId(ip, port, profileType)
		err = newProfile(profileId, ip, port, seconds, profileType)
		if err != nil {
			rsp.WriteHeader(http.StatusInternalServerError)
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
<div>
	<label for="ip">IP:</label><br>
	<input type="text" id="ip" name="ip" size="20" placeholder="10.0.0.1" autofocus required><br>
</div>
<div>
	<label for="port">Port:</label><br>
	<input type="number" id="port" name="port" size="20" placeholder="8000" min="1" max="65535" required><br>
</div>
<div>
	<label for="seconds">Seconds:</label><br>
	<input type="number" id="seconds" name="seconds" size="20" placeholder="30" min="1" max="60" required><br>
</div>
<fieldset>
	<legend>Select profile type:</legend>
	<div>
		<input type="radio" id="cpu" name="type" value="cpu" checked>
		<label for="cpu">CPU</label>
	</div>
	<div>
		<input type="radio" id="heap" name="type" value="heap">
		<label for="heap">heap</label>
	</div>
	<div>
		<input type="radio" id="goroutine" name="type" value="goroutine">
		<label for="goroutine">goroutine</label>
	</div>
</fieldset>
<br>
<div>
	<input type="submit" value="Profile!" onClick="this.form.submit(); this.disabled=true; this.value='Profiling…';">
</div>
</form>
</body>
</html>`))
	return
}

func handleProfile(rsp http.ResponseWriter, req *http.Request, profileId string, pathHandle map[string]http.Handler) {
	realPath := req.URL.Path[len(profileId)+1:]
	if realPath == "" {
		req.URL.Path += "/"
		http.Redirect(rsp, req, req.URL.String(), http.StatusFound)
		return
	}
	var handle http.Handler
	var ok bool
	if handle, ok = pathHandle[realPath]; !ok {
		handle = pathHandle["/"]
	}
	handle.ServeHTTP(rsp, req)
}

func newProfile(profileId, ip string, port, seconds int, profileType ProfileType) error {
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
			args:    []string{profilePath},
		},
		Sym:        &webSym{Address: ip + ":" + strconv.Itoa(port)},
		UI:         &webUI{},
		HTTPServer: pprofHTTPServer,
	}
}

func fetchProfile(profileId, ip string, port, seconds int, profileType ProfileType) (string, error) {
	var typePart string
	switch profileType {
	case ProfileTypeCPU:
		typePart = "profile"
		break
	case ProfileTypeHeap:
		typePart = "heap"
		break
	case ProfileTypeGoroutine:
		typePart = "goroutine"
		break
	default:
		return "", errors.New("unknown type")
	}
	url := fmt.Sprintf("http://%s:%d/debug/pprof/%s?seconds=%d", ip, port, typePart, seconds)

	client := &http.Client{
		Timeout: time.Duration(seconds)*time.Second + 5*time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Println("http fetch: ", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("http fetch get status: ", resp.StatusCode)
		return "", fmt.Errorf("wrong http status: %d", resp.StatusCode)
	}

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

func newProfileId(ip string, port int, profileType ProfileType) string {
	u := uuid.NewString()
	profileId := ip + "_" + strconv.Itoa(port) + "_" + u
	switch profileType {
	case ProfileTypeCPU:
		return profileId + "_cpu"
	case ProfileTypeHeap:
		return profileId + "_heap"
	case ProfileTypeGoroutine:
		return profileId + "_goroutine"
	default:
		return profileId + "_unknown"
	}
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
