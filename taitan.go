package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/datasektionen/taitan/pages"
	"golang.org/x/exp/inotify"
)

var (
	debug     bool   // Show debug level messages.
	info      bool   // Show info level messages.
	responses Atomic // Our parsed responses.
)

func init() {
	flag.BoolVar(&debug, "vv", false, "Print debug messages.")
	flag.BoolVar(&info, "v", false, "Print info messages.")
	flag.Usage = usage
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] ROOT\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

// setVerbosity sets the amount of messages printed.
func setVerbosity() {
	if debug {
		log.SetLevel(log.DebugLevel)
		return
	}
	if info {
		log.SetLevel(log.InfoLevel)
		return
	}
	log.SetLevel(log.WarnLevel)
}

type Atomic struct {
	sync.Mutex
	Resps map[string]*pages.Resp
}

func main() {
	setVerbosity()

	// We need a root folder.
	if flag.NArg() < 1 {
		usage()
	}

	// Recieve port from heruko.
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("$PORT environmental variable is not set.")
	}

	// Our root to read in markdown files.
	root := flag.Arg(0)
	log.WithField("Root", root).Info("Our root directory")

	// We'll parse and store the responses ahead of time.
	resps, err := pages.Load(root)
	if err != nil {
		log.Fatalf("loadRoot: unexpected error: %#v\n", err)
	}
	log.WithField("Resps", resps).Debug("The parsed responses")
	responses = Atomic{Resps: resps}
	go watch(root)

	log.Info("Starting server.")
	log.Info("Listening on port: ", port)

	// Our request handler.
	http.HandleFunc("/", handler)

	// Listen on port and serve with our handler.
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func watch(root string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(wd, root)
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.AddWatch(path,
		inotify.IN_CLOSE_WRITE|
			inotify.IN_CREATE|
			inotify.IN_DELETE|
			inotify.IN_MODIFY|
			inotify.IN_MOVED_FROM|
			inotify.IN_MOVED_TO|
			inotify.IN_MOVE)
	if err != nil {
		log.Fatal(err)
	}
	last := time.Now()
	for {
		select {
		case ev := <-watcher.Event:
			if time.Now().Sub(last) < 10*time.Second {
				continue
			}
			log.Println("event:", ev)
			last = time.Now()
			responses.Lock()
			responses.Resps, err = pages.Load(root)
			if err != nil {
				log.Error(err)
			}
			responses.Unlock()
		case err := <-watcher.Error:
			log.Warn("error:", err)
		}
	}
}

// handler parses and serves responses to our file queries.
func handler(res http.ResponseWriter, req *http.Request) {
	// Requested URL. We extract the path.
	query := req.URL.Path
	log.WithField("query", query).Info("Recieved query")

	clean := filepath.Clean(query)
	log.WithField("clean", clean).Info("Sanitized path")

	r, ok := responses.Resps[clean]
	if !ok {
		log.WithField("page", clean).Warn("Page doesn't exist")
		res.WriteHeader(404)
		return
	}
	log.Info("Marshaling the response.")
	buf, err := json.Marshal(r)
	if err != nil {
		log.Warnf("handler: unexpected error: %#v\n", err)
		res.WriteHeader(500)
		return
	}
	log.Info("Serve the response.")
	log.Debug("Response: %#v\n", string(buf))
	fmt.Fprintln(res, string(buf))
}
