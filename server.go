package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const ()

type Configs struct {
	ServerFiles      []string
	AutoTypes  []string
	ServerAddr string
	ServerPath string
}

var (
	serverAddr     = ":9000" // change to your serveraddress and port
	serverConfigs  Configs
	serverFileList []string
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

func main() {
	if len(serverConfigs.ServerPath) > 0 {
		os.Chdir(serverConfigs.ServerPath)
	}
	wd, _ := os.Getwd()
	log.Println("cd ->", serverConfigs.ServerPath)
	log.Println("Server direktory:", wd)
	
	// make ServerFile list
	if len(serverConfigs.ServerFiles) == 0 && len(serverConfigs.AutoTypes) > 0 {
		lsDir(".")
		for _, fNme := range serverFileList {
			for _, ending := range serverConfigs.AutoTypes {
				if strings.EqualFold(ending, fNme[len(fNme)-len(ending):]) {
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
	log.Fatal(serv.ListenAndServe())

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
	log.Println(r.RequestURI)
	req := strings.SplitAfterN(r.RequestURI, `/`, 2)
	if len(req[1]) < 2 {
		http.ServeFile(w, r, "index.html")
		return
	}

	if len(serverConfigs.ServerFiles) > 0 {
		for _, f := range serverConfigs.ServerFiles {
			if strings.EqualFold(req[1], f) {
				http.ServeFile(w, r, req[1])
				return
			}
		}
		w.WriteHeader(418)
		return
	}

	http.ServeFile(w, r, req[1])
}
