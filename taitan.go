package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/rjeczalik/notify"
	"github.com/datasektionen/taitan/fuzz"
	"github.com/datasektionen/taitan/pages"
)

var (
	debug     bool   // Show debug level messages.
	info      bool   // Show info level messages.
	watch     bool   // Watch for file changes.
	responses Atomic // Our parsed responses.
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.BoolVar(&debug, "vv", false, "Print debug messages.")
	flag.BoolVar(&info, "v", false, "Print info messages.")
	flag.BoolVar(&watch, "w", false, "Watch for file changes.")
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
		runGit("clone", []string{"clone", u.String()})
	} else {
		runGit("pull", []string{"-C", root, "pull"})
	}
}

func runGit(action string, args []string) {
	log.Infof("Found root directory - %sing updates!", action)
	log.Debugf("Commands %#v!", args)
	cmd := exec.Command("git", args...)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Waiting for git %s to finish...", action)
	err = cmd.Wait()
	if err != nil {
		log.Warnf("%sed with error: %v\n", action, err)
	}
	log.Infof("Git %s finished!", action)
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

var jumpfile map[string]interface{}

func main() {
	setVerbosity()

	// Get port or die.
	port := getEnv("PORT")

	// Get content or die.
	getContent()

	root := getRoot()
	log.WithField("Root", root).Info("Our root directory")

	// We'll parse and store the responses ahead of time.
	resps, err := pages.Load(root)
	if err != nil {
		log.Fatalf("pages.Load: unexpected error: %s", err)
	}
	log.WithField("Resps", resps).Debug("The parsed responses")
	responses = Atomic{Resps: resps}

	log.Info("Starting server.")
	log.Info("Listening on port: ", port)

	// If jumpfile exists.
	if _, err := os.Stat(root + "/jumpfile.json"); err == nil {
		buf, err := ioutil.ReadFile(root + "/jumpfile.json")
		if err != nil {
			log.Fatalf("jumpfile readfile: unexpected error: %s", err)
		}
		err = json.Unmarshal(buf, &jumpfile)
		if err != nil {
			log.Fatalf("jumpfile unmarshal: unexpected error: %s", err)
		}
		log.Debugln(jumpfile)
	} else {
		log.Infoln("No jumpfile found")
	}

	// Our request handler.
	http.HandleFunc("/", handler)

	if watch {
		events := make(chan notify.EventInfo, 5)

		if err := notify.Watch(fmt.Sprintf("%s/...", root),
								events,
								notify.Create,
								notify.Remove,
								notify.Write,
								notify.Rename); err != nil {
			log.Fatal(err)
		}
		defer notify.Stop(events)

		go func() {
			for event := range events {
				log.Info("event:", event)
				// Lacks error handling as this should not run in production.
				responses.Resps, _ = pages.Load(getRoot())
			}
		}()
	}

	// Listen on port and serve with our handler.
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

// handler parses and serves responses to our file queries.
func handler(res http.ResponseWriter, req *http.Request) {
	if v, ok := jumpfile[filepath.Clean(req.URL.Path)]; ok {
		req.URL.Path = v.(string)
		http.Redirect(res, req, req.URL.String(), http.StatusSeeOther)
		log.Infoln("Redirect: " + req.URL.String())
		return
	}
	if req.URL.Path == "/fuzzyfile" {
		log.Info("Fuzzyfile")
		buf, err := json.Marshal(fuzz.NewFile(responses.Resps))
		if err != nil {
			log.Warnf("handler: unexpected error: %#v\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf("Response: %#v\n", string(buf))
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Write(buf)
		return
	}
	if req.Header.Get("X-Github-Event") == "push" {
		var err error
		log.Infoln("Push hook")
		getContent()
		responses.Lock()
		responses.Resps, err = pages.Load(getRoot())
		if err != nil {
			log.Error(err)
		}
		responses.Unlock()
		return
	}

	// Requested URL. We extract the path.
	query := req.URL.Path
	log.WithField("query", query).Info("Recieved query")

	clean := filepath.Clean(query)
	log.WithField("clean", clean).Info("Sanitized path")
	log.Println(rootDir(clean))
	responses.Lock()
	r, ok := responses.Resps[clean]
	responses.Unlock()
	if !ok {
		log.WithField("page", clean).Warn("Page doesn't exist")
		res.WriteHeader(http.StatusNotFound)
		return
	}

	// Sort the slugs
	var slugs []string
	for k := range responses.Resps {
		slugs = append(slugs, k)
	}
	sort.Strings(slugs)

	// Our web tree.
	root := pages.NewNode("/", "/", responses.Resps["/"].Title)
	for _, slug := range slugs {
		// if strings.HasPrefix(slug, filepath.Dir(clean)) {
		root.AddNode(
			strings.FieldsFunc(clean, func(c rune) bool { return c == '/' }),
			slug,
			responses.Resps[slug].Title,
			strings.FieldsFunc(slug, func(c rune) bool { return c == '/' }),
			false,
			false,
		)
		// }
	}
	r.URL = clean
	if root.Num() == 1 {
		r.Nav = nil
	} else {
		r.Nav = root.Nav
	}

	log.Info("Marshaling the response.")
	buf, err := json.Marshal(r)
	if err != nil {
		log.Warnf("handler: unexpected error: %#v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Serve the response.")
	log.Debugf("Response: %#v\n", string(buf))
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(buf)
}

func rootDir(path string) string {
	for {
		test := filepath.Dir(path)
		if test == "." || test == "/" {
			break
		}
		path = test
	}
	return path
}
