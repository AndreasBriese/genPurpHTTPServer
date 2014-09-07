package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const ()

type Configs struct {
	ServerFiles []string
	AutoTypes   []string
	ServerAddr  string
	ServerPath  string
	ServerLogging bool
	Tarpit      bool
}

var (
	serverAddr            = ":9000" // change to your serveraddress and port
	serverConfigs         Configs
	serverFileList        []string
	TarpitHTTPStatusCodes = []int{
		200, 201, 202, 203, 206, 207,
		300, 305, 306,
		400, 401, 402, 403, 404, 405, 406, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 420, 422, 423, 424, 425, 426,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 509, 510,
	}
)

func init() {
	env := os.Environ()
	for _, evs := range env {
		e := strings.SplitAfterN(evs, "=", 2)
		// log.Println(evs, e)
		if e[0] == "_=" {
			ee := strings.Split(e[1], string(os.PathSeparator))
			os.Chdir(strings.Join(ee[0:len(ee)-1], string(os.PathSeparator)))
		}
	}
	serverConfigs = loadConfigs()
	if len(serverConfigs.ServerAddr) > 0 {
		serverAddr = serverConfigs.ServerAddr
	}
	
	serverAddr = *(flag.String("address", serverAddr, "Server address"))
	flag.Parse()
}

func main() {
	if len(serverConfigs.ServerPath) > 0 {
		os.Chdir(serverConfigs.ServerPath)
	}
	wd, _ := os.Getwd()
	log.Println("cd ->", serverConfigs.ServerPath)
	log.Println("Server dir:", wd)

	// make ServerFile list
	if len(serverConfigs.ServerFiles) == 0 && len(serverConfigs.AutoTypes) > 0 {
		lsDir(".")
		for _, fNme := range serverFileList {
			for _, ending := range serverConfigs.AutoTypes {
				if len(fNme) > len(ending)+2 && strings.EqualFold(ending, fNme[len(fNme)-len(ending):]) {
					serverConfigs.ServerFiles = append(serverConfigs.ServerFiles, fNme[2:])
				}
			}
		}
	}
	log.Println("Server will serve these files:", serverConfigs.ServerFiles)

	// start Server
	http.HandleFunc("/", logPanic(rootHandler))
	serv := &http.Server{
		Addr:           serverAddr,
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}
	
	log.Println("Server starts at", serverAddr)
	
	if !serverConfigs.ServerLogging {
		log.Println("Config says: No server logging")
	}
	
	if serverConfigs.Tarpit {
		log.Println("Spider/pentester tarpit activated!")
	}
	
	log.Fatal(serv.ListenAndServe())

}

func lsDir(dir string) {
	ls, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
	}
	for _, f := range ls {
		if f.Name()[0] == '.' || f.Name() == ".DS_Store" {
			continue
		}
		fleNme := dir + string(os.PathSeparator) + f.Name()
		// log.Printf("ls: %v", fleNme)
		serverFileList = append(serverFileList, fleNme)
		if f.IsDir() {
			lsDir(fleNme)
		}
	}
}

func loadConfigs() (cfg Configs) {
	cfgFle, err := os.Open("./SERVER.conf")
	if err != nil {
		return cfg
	}
	cfgs, err := ioutil.ReadAll(cfgFle)
	if err != nil {
		return cfg
	}
	json.Unmarshal(cfgs, &cfg)
	return cfg
}

func logPanic(function func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				log.Printf("[%v] caught panic: %v", r.RemoteAddr, x)
			}
		}()
		function(w, r)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if serverConfigs.ServerLogging {
		log.Printf("(%v) %v \"%s %s %s\" \"%s\" \"%s\"\n",
			time.Now().Unix(),
			r.RemoteAddr,
			r.Method,
			r.URL.String(),
			r.Proto,
			r.Referer(),
			r.UserAgent(),
		)
	}

	req := strings.SplitAfterN(r.RequestURI, `/`, 2)

	if len(req[1]) < 2 {
		req[1] = "indEX.html"
	}

	if len(serverConfigs.ServerFiles) > 0 {
		for _, f := range serverConfigs.ServerFiles {
			if strings.EqualFold(req[1], f) {
				http.ServeFile(w, r, f)
				return
			}
		}
		if serverConfigs.Tarpit {
			w.WriteHeader(TarpitHTTPStatusCodes[rand.Intn(len(TarpitHTTPStatusCodes))])
			return
		} else {
			w.WriteHeader(404)
			return
		}

	}

	http.ServeFile(w, r, req[1])
}
