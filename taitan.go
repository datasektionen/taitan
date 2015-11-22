package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] ROOT\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&debug, "vv", false, "Print debug messages.")
	flag.BoolVar(&info, "v", false, "Print info messages.")
	flag.Usage = usage
	flag.Parse()
}

func getEnv(env string) string {
	e := os.Getenv(env)
	if e == "" {
		log.Fatalf("$%s environmental variable is not set.\n", env)
	}
	return e
}

func getRoot() string {
	content := getEnv("CONTENT_URL")
	u, err := url.Parse(content)
	if err != nil {
		log.Fatalln("getContent: ", err)
	}

	// https://<token>@github.com/username/repo.git
	u.User = url.User(getEnv("TOKEN"))

	base := filepath.Base(u.Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func getContent() {
	content := getEnv("CONTENT_URL")
	u, err := url.Parse(content)
	if err != nil {
		log.Fatalln("getContent: ", err)
	}

	// https://<token>@github.com/username/repo.git
	u.User = url.User(getEnv("TOKEN"))

	root := getRoot()
	if _, err = os.Stat(root); os.IsNotExist(err) {
		log.Debugln("No root directory - cloning content url!")
		cmd := exec.Command("git", "clone", u.String())
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Waiting for git clone to finish...")
		err = cmd.Wait()
		if err != nil {
			log.Warnln("Cloned with error: %v\n", err)
		}
	} else {
		log.Debugln("Found root directory - pulling updates!")
		cmd := exec.Command("git", fmt.Sprintf("--git-dir=%s/.git", root), "pull")
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Waiting for git pull to finish...")
		err = cmd.Wait()
		if err != nil {
			log.Warnf("Pulled with error: %v\n", err)
		}
	}
}

// setVerbosity sets the amount of messages printed.
func setVerbosity() {
	switch {
	case debug:
		log.SetLevel(log.DebugLevel)
	case info:
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

// Atomic responses.
type Atomic struct {
	sync.Mutex
	Resps map[string]*pages.Resp
}

func validRoot(root string) {
	fi, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Directory doesn't exist: %q", root)
		}
		log.Fatalln(err)
	}
	if !fi.IsDir() {
		log.Fatalf("Supplied path is not a directory: %q", root)
	}
}

func main() {
	setVerbosity()

	// Get port or die.
	port := getEnv("PORT")

	// Get content or die.
	getContent()
	go func() {
		for {
			time.Sleep(time.Second * 20)
			getContent()
		}
	}()

	root := getRoot()
	log.WithField("Root", root).Info("Our root directory")

	// We'll parse and store the responses ahead of time.
	resps, err := pages.Load(root)
	if err != nil {
		log.Fatalf("pages.Load: unexpected error: %s", err)
	}
	log.WithField("Resps", resps).Debug("The parsed responses")
	responses = Atomic{Resps: resps}

	// Watch the directory for any changes. If the directory has any changes we'll
	// update our responses.
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

	responses.Lock()
	r, ok := responses.Resps[clean]
	responses.Unlock()
	if !ok {
		log.WithField("page", clean).Warn("Page doesn't exist")
		res.WriteHeader(http.StatusNotFound)
		return
	}
	log.Info("Marshaling the response.")
	buf, err := json.Marshal(r)
	if err != nil {
		log.Warnf("handler: unexpected error: %#v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Serve the response.")
	log.Debug("Response: %#v\n", string(buf))
	res.Header().Set("Content-Type", "application/json")
	res.Write(buf)
}
