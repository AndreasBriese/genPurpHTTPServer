// The MIT License (MIT)

// Copyright (c) 2014 Andreas Briese, eduToolbox@Bri-C GmbH, Sarstedt GERMANY

//
// tarpit inspired by https://www.youtube.com/watch?v=I3pNLB3Cq24
//
// indicated loc from the go/golang developers
//

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package main

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const ()

type Configs struct {
	ServerFiles   []string
	AutoTypes     []string
	ServerAddr    string
	ServerPath    string
	ServerLogging bool
	RunTarpit     bool
}

var (
	serverAddr               = ":9000" // change to your default serveraddress and port
	serverConfigs            Configs
	serverFileList           []string
	RunTarpitHTTPStatusCodes = []int{
		200, 201, 202, 203, 206, 207,
		300, 305, 306,
		400, 401, 402, 403, 404, 405, 406, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 420, 422, 423, 424, 425, 426,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 509, 510,
	}
)

func init() {
	env := os.Environ()
	var home string
	for _, evs := range env {
		e := strings.SplitAfterN(evs, "=", 2)
		// log.Println(evs, e)
		switch e[0] {
		case "_=":
			ee := strings.Split(e[1], string(os.PathSeparator))
			os.Chdir(strings.Join(ee[0:len(ee)-1], string(os.PathSeparator)))
		case "HOME=":
			home = e[1]
		}
	}
	serverConfigs = loadConfigs()
	// log.Println(serverConfigs)
	if len(serverConfigs.ServerAddr) > 0 {
		serverAddr = serverConfigs.ServerAddr
	}
	if len(serverConfigs.ServerPath) > 0 && serverConfigs.ServerPath[0] == '~' {
		serverConfigs.ServerPath = home + serverConfigs.ServerPath[1:]
		os.Chdir(serverConfigs.ServerPath)
	}

	// eventually override serverAddr by commandline option
	flag.StringVar(&serverAddr, "address", serverAddr, "Server address")
}

func main() {
	flag.Parse()
	if len(serverConfigs.ServerPath) > 0 {
		os.Chdir(serverConfigs.ServerPath)
	}
	if wd, err := os.Getwd(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("cd ->", serverConfigs.ServerPath)
		log.Println("Server dir:", wd)
	}

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

	// check files on ServerFile list
	serverFileList = []string{}
	for _, fNme := range serverConfigs.ServerFiles {
		// check files on list
		fstat, err := os.Stat(fNme)
		if err != nil || fstat.IsDir() || fstat.Mode()&^07777 == os.ModeSocket {
			log.Printf("file '%v' is not available \n", fNme)
			continue
		}
		serverFileList = append(serverFileList, fNme)
	}

	log.Println("Server will serve these files:")
	serverConfigs.ServerFiles = []string{}
	for _, fNme := range serverFileList {
		serverConfigs.ServerFiles = append(serverConfigs.ServerFiles, fNme)
		log.Println("  ", fNme)
	}

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

	if serverConfigs.RunTarpit {
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
		if f.Name()[0] == '.' {
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
	log.Println("./SERVER.conf loaded")
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

func compressedServe(w http.ResponseWriter, r *http.Request, filePath string, out io.Writer) bool {
	f, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(404)
		return true
	}
	defer f.Close()

	fstat, err := f.Stat()

	if err != nil {
		w.WriteHeader(500)
		return true
	}

	if fstat.Mode()&^07777 == os.ModeSocket {
		// don't give access 2 sockets !
		w.WriteHeader(403)
		return true
	}

	// from Golang http.ServeFile code
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && fstat.ModTime().Unix() <= t.Unix() {
		w.WriteHeader(304)
		return true
	}
	w.Header().Set("Last-Modified", fstat.ModTime().Format(http.TimeFormat))

	// set header 4 mimetype
	if mimeTyp := mime.TypeByExtension(path.Ext(filePath)); mimeTyp != "" {
		mimeCat := strings.Split(mimeTyp, "/")
		if mimeCat[0] == "video" || mimeCat[0] == "audio" || mimeCat[1] == "pdf" {
			// leave AV content to http.ServeFile
			return false
		}
		w.Header().Set("Content-Type", mimeTyp)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	var buf = make([]byte, fstat.Size())
	var n int
	for err == nil {
		n, err = f.Read(buf)
		out.Write(buf[0:n])
	}

	switch out.(type) {
	case *gzip.Writer:
		out.(*gzip.Writer).Close()
	case *zlib.Writer:
		out.(*zlib.Writer).Close()
	}

	return true
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
				//
				// file is in list 4 serving
				if strings.EqualFold(f[len(f)-4:], ".jpg") {
					// serve jpgs uncompressed
					http.ServeFile(w, r, f)
					return
				}
				var servedCompressed bool
				if r.Header.Get("Accept-Encoding") != "" {
					// client accepts encoding
					// check 4 gzip / deflate
				loop:
					for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
						switch enc {
						case "gzip":
							w.Header().Set("Content-Encoding", "gzip")
							out := gzip.NewWriter(w)
							servedCompressed = compressedServe(w, r, f, out)
							break loop
						case "deflate":
							w.Header().Set("Content-Encoding", "deflate")
							out := zlib.NewWriter(w)
							servedCompressed = compressedServe(w, r, f, out)
							break loop
						}
					}
				}
				// else serve uncompressed
				if !servedCompressed {
					http.ServeFile(w, r, f)
				}
				return
			}
		}
		if serverConfigs.RunTarpit {
			tarpitHeader := RunTarpitHTTPStatusCodes[rand.Intn(len(RunTarpitHTTPStatusCodes))]
			log.Println("throw tar -> HTTP Status", tarpitHeader)
			w.WriteHeader(tarpitHeader)
			return
		} else {
			w.WriteHeader(404)
			return
		}

	}

	http.ServeFile(w, r, req[1])
}
