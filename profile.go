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
			_, _ = rsp.Write([]byte("invalid ip"))
			return
		}

		var port int
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			rsp.WriteHeader(http.StatusBadRequest)
			_, _ = rsp.Write([]byte("invalid port"))
			return
		}

		var seconds int
		if secondsStr == "" {
			seconds = 30
		} else {
			seconds, err = strconv.Atoi(secondsStr)
			if err != nil {
				rsp.WriteHeader(http.StatusBadRequest)
				_, _ = rsp.Write([]byte("invalid seconds"))
				return
			}
			if seconds < 1 {
				seconds = 1
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
			_, _ = rsp.Write([]byte("wrong profile type"))
			return
		}
		profileId := newProfileId(ip, port, profileType)
		err = newProfile(profileId, ip, port, seconds, profileType)
		if err != nil {
			rsp.WriteHeader(http.StatusInternalServerError)
			_, _ = rsp.Write([]byte("fetch failed.\n" + err.Error()))
			return
		}
		http.Redirect(rsp, req, buildPathFromBase("./"+profileId+"/"), http.StatusFound)
		return
	}

	_, _ = rsp.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PProf-Web</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-900 min-h-screen flex items-center justify-center p-4">
    <div class="w-full max-w-lg">
        <!-- Header -->
        <div class="text-center mb-8">
            <div class="inline-flex items-center justify-center mb-4">
                <svg class="w-10 h-10 text-cyan-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.148 2.148A12.061 12.061 0 0116.5 7.605"/>
                </svg>
            </div>
            <h1 class="text-2xl font-bold text-white">PProf-Web</h1>
            <p class="text-slate-400 mt-1 text-sm">Fetch and visualize Go pprof profiles through a web interface</p>
        </div>

        <!-- Card -->
        <div class="bg-slate-800 rounded-xl shadow-2xl border border-slate-700 p-6">
            <form id="profileForm" onsubmit="return handleSubmit()">
                <!-- IP and Port row -->
                <div class="grid grid-cols-3 gap-3 mb-4">
                    <div class="col-span-2">
                        <label for="ip" class="block text-xs font-medium text-slate-400 mb-1.5">IP Address</label>
                        <input type="text" id="ip" name="ip" placeholder="10.0.0.1" required
                            class="w-full px-3 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition">
                    </div>
                    <div>
                        <label for="port" class="block text-xs font-medium text-slate-400 mb-1.5">Port</label>
                        <input type="number" id="port" name="port" placeholder="8000" required min="1" max="65535"
                            class="w-full px-3 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition">
                    </div>
                </div>

                <!-- Seconds -->
                <div class="mb-5">
                    <label for="seconds" class="block text-xs font-medium text-slate-400 mb-1.5">Duration (seconds)</label>
                    <input type="number" id="seconds" name="seconds" placeholder="30" min="1" max="60"
                        class="w-full px-3 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition">
                    <p class="text-slate-500 text-xs mt-1">Used for CPU profiles. Max 60 seconds.</p>
                </div>

                <!-- Profile Type Cards -->
                <fieldset class="mb-5">
                    <legend class="block text-xs font-medium text-slate-400 mb-2">Profile Type</legend>
                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                        <label class="relative cursor-pointer">
                            <input type="radio" name="type" value="cpu" class="peer sr-only" checked>
                            <div class="p-3 rounded-lg border border-slate-600 bg-slate-900 peer-checked:border-cyan-500 peer-checked:bg-cyan-500/10 hover:border-slate-500 transition">
                                <div class="text-sm font-semibold text-white">CPU</div>
                                <div class="text-xs text-slate-400 mt-0.5">Function time spent</div>
                            </div>
                        </label>
                        <label class="relative cursor-pointer">
                            <input type="radio" name="type" value="heap" class="peer sr-only">
                            <div class="p-3 rounded-lg border border-slate-600 bg-slate-900 peer-checked:border-cyan-500 peer-checked:bg-cyan-500/10 hover:border-slate-500 transition">
                                <div class="text-sm font-semibold text-white">Heap</div>
                                <div class="text-xs text-slate-400 mt-0.5">Memory allocations</div>
                            </div>
                        </label>
                        <label class="relative cursor-pointer">
                            <input type="radio" name="type" value="goroutine" class="peer sr-only">
                            <div class="p-3 rounded-lg border border-slate-600 bg-slate-900 peer-checked:border-cyan-500 peer-checked:bg-cyan-500/10 hover:border-slate-500 transition">
                                <div class="text-sm font-semibold text-white">Goroutine</div>
                                <div class="text-xs text-slate-400 mt-0.5">Stack traces</div>
                            </div>
                        </label>
                    </div>
                </fieldset>

                <!-- Submit -->
                <button type="submit" id="submitBtn"
                    class="w-full px-4 py-2.5 text-sm font-medium text-white bg-cyan-600 rounded-lg hover:bg-cyan-500 focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-2 focus:ring-offset-slate-800 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                    <span id="btnText">Fetch Profile</span>
                    <svg id="spinner" class="hidden animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                </button>
            </form>
        </div>

        <!-- History -->
        <div id="historySection" class="mt-4 hidden">
            <div class="bg-slate-800 rounded-xl border border-slate-700 p-4">
                <div class="flex items-center justify-between mb-3">
                    <h2 class="text-xs font-medium text-slate-400">Recent Profiles</h2>
                    <button onclick="clearHistory()" class="text-xs text-slate-500 hover:text-red-400 transition">Clear all</button>
                </div>
                <div id="historyList" class="space-y-1.5"></div>
            </div>
        </div>

        <!-- Footer -->
        <p class="text-center text-slate-600 text-xs mt-6">
            Connects to the target's /debug/pprof endpoint to fetch the profile.
        </p>
    </div>

    <script>
    var STORAGE_KEY = "pprof-web-history";
    var MAX_HISTORY = 20;

    function getHistory() {
        try {
            return JSON.parse(localStorage.getItem(STORAGE_KEY)) || [];
        } catch(e) {
            return [];
        }
    }

    function saveHistory(list) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(list));
    }

    function addToHistory(ip, port, seconds, type) {
        var list = getHistory();
        var key = ip + ":" + port + ":" + type + ":" + seconds;
        list = list.filter(function(item) {
            return (item.ip + ":" + item.port + ":" + item.type + ":" + item.seconds) !== key;
        });
        list.unshift({ip: ip, port: port, seconds: seconds, type: type, ts: Date.now()});
        if (list.length > MAX_HISTORY) list = list.slice(0, MAX_HISTORY);
        saveHistory(list);
    }

    function removeFromHistory(index) {
        var list = getHistory();
        list.splice(index, 1);
        saveHistory(list);
        renderHistory();
    }

    function clearHistory() {
        saveHistory([]);
        renderHistory();
    }

    function pickHistory(index) {
        var list = getHistory();
        var item = list[index];
        if (!item) return;
        document.getElementById("ip").value = item.ip;
        document.getElementById("port").value = item.port;
        document.getElementById("seconds").value = item.seconds || "";
        var radios = document.querySelectorAll("input[name=type]");
        for (var i = 0; i < radios.length; i++) {
            radios[i].checked = (radios[i].value === item.type);
        }
    }

    function typeLabel(t) {
        if (t === "cpu") return "CPU";
        if (t === "heap") return "Heap";
        if (t === "goroutine") return "Goroutine";
        return t;
    }

    function renderHistory() {
        var list = getHistory();
        var section = document.getElementById("historySection");
        var container = document.getElementById("historyList");
        if (list.length === 0) {
            section.classList.add("hidden");
            return;
        }
        section.classList.remove("hidden");
        var html = "";
        for (var i = 0; i < list.length; i++) {
            var item = list[i];
            var sec = item.seconds ? item.seconds + "s" : "30s";
            html += "<div class=\"flex items-center gap-2 group\">"
                + "<button type=\"button\" onclick=\"pickHistory(" + i + ")\" "
                + "class=\"flex-1 text-left px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 hover:border-cyan-500/50 transition text-sm\">"
                + "<span class=\"text-white\">" + item.ip + ":" + item.port + "</span>"
                + "<span class=\"ml-2 text-slate-500\">" + typeLabel(item.type) + "</span>"
                + "<span class=\"ml-1 text-slate-600\">" + sec + "</span>"
                + "</button>"
                + "<button type=\"button\" onclick=\"removeFromHistory(" + i + ")\" "
                + "class=\"p-1.5 rounded text-slate-600 hover:text-red-400 opacity-0 group-hover:opacity-100 transition\" title=\"Remove\">"
                + "<svg class=\"w-4 h-4\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" viewBox=\"0 0 24 24\">"
                + "<path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M6 18L18 6M6 6l12 12\"/>"
                + "</svg></button></div>";
        }
        container.innerHTML = html;
    }

    function handleSubmit() {
        var ip = document.getElementById("ip").value;
        var port = document.getElementById("port").value;
        var seconds = document.getElementById("seconds").value;
        var type = document.querySelector("input[name=type]:checked").value;
        addToHistory(ip, port, seconds, type);

        var btn = document.getElementById("submitBtn");
        var txt = document.getElementById("btnText");
        var spin = document.getElementById("spinner");
        btn.disabled = true;
        txt.textContent = "Fetching profile...";
        spin.classList.remove("hidden");
        return true;
    }

    renderHistory();
    </script>
</body>
</html>`))
	return
}

func handleProfile(rsp http.ResponseWriter, req *http.Request, profileId string, pathHandle map[string]http.Handler) {
	absPath := getPathFromBase(req.URL.Path)
	realPath := absPath[len(profileId)+1:]
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
		if seconds <= 0 {
			seconds = 30
		}
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
	var url string
	if seconds == 0 {
		url = fmt.Sprintf("http://%s:%d/debug/pprof/%s", ip, port, typePart)
	} else {
		url = fmt.Sprintf("http://%s:%d/debug/pprof/%s?seconds=%d", ip, port, typePart, seconds)
	}

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
	defer f.Close()
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
