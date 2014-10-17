//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package main

import (
	"expvar"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	"github.com/gorilla/mux"

	bleveHttp "github.com/blevesearch/bleve/http"

	log "github.com/couchbaselabs/clog"
)

var bindAddr = flag.String("addr", ":8095",
	"http listen [address]:port")
var dataDir = flag.String("dataDir", "data",
	"directory for index data")
var logFlags = flag.String("logFlags", "",
	"comma-separated clog flags, to control logging")
var staticDir = flag.String("staticDir", "static",
	"directory for static web UI content")
var staticEtag = flag.String("staticEtag", "",
	"static etag value")
var server = flag.String("server", "",
	"url to couchbase server, example: http://localhost:8091")

var expvars = expvar.NewMap("stats")

func init() {
	expvar.Publish("cbft", expvars)
	expvars.Set("indexes", bleveHttp.IndexStats())
}

func main() {
	flag.Parse()

	go dumpOnSignalForPlatform()

	log.Printf("%s started", os.Args[0])
	if *logFlags != "" {
		log.ParseLogFlag(*logFlags)
	}
	flag.VisitAll(func(f *flag.Flag) { log.Printf("  -%s=%s\n", f.Name, f.Value) })
	log.Printf("  GOMAXPROCS=%d", runtime.GOMAXPROCS(-1))

	router, err := mainStart(*dataDir, *staticDir, *server)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", router)
	log.Printf("listening on: %v", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}

func mainStart(dataDir, staticDir, server string) (*mux.Router, error) {
	if server == "" {
		return nil, fmt.Errorf("error: couchbase server URL required (-server)")
	}

	mgr := NewManager(dataDir, server, &MainHandlers{})
	err := mgr.Start()
	if err != nil {
		return nil, err
	}

	return NewManagerHTTPRouter(mgr, staticDir)
}

type MainHandlers struct{}

func (meh *MainHandlers) OnRegisterPIndex(pindex *PIndex) {
	bleveHttp.RegisterIndexName(pindex.name, pindex.BIndex())
}

func (meh *MainHandlers) OnUnregisterPIndex(pindex *PIndex) {
	bleveHttp.UnregisterIndexByName(pindex.name)
}

func dumpOnSignal(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	for _ = range c {
		pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
	}
}