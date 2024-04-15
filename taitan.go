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
	"sort"
	"strings"
	"sync"

	"github.com/datasektionen/taitan/fuzz"
	"github.com/datasektionen/taitan/pages"
	"github.com/rjeczalik/notify"
	log "github.com/sirupsen/logrus"
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
	e, found := os.LookupEnv(env)
	if !found {
		log.Fatalf("$%s environmental variable is not set.\n", env)
	}
	return e
}

func getRoot() string {
	if contentDir, ok := os.LookupEnv("CONTENT_DIR"); ok {
		return contentDir
	}
	content := getEnv("CONTENT_URL")
	u, err := url.Parse(content)
	if err != nil {
		log.Fatalln("getContent: ", err)
	}

	base := filepath.Base(u.Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func getContent() error {
	if _, ok := os.LookupEnv("CONTENT_DIR"); ok {
		return nil
	}
	content := getEnv("CONTENT_URL")
	u, err := url.Parse(content)
	if err != nil {
		log.Fatalln("getContent: ", err)
	}

	// https://<token>@github.com/username/repo.git
	githubToken, tokenFound := os.LookupEnv("TOKEN")
	if tokenFound {
		u.User = url.User(githubToken)
	}

	root := getRoot()
	if _, err = os.Stat(root); os.IsNotExist(err) {
		if err := runGit("clone", "clone", u.String()); err != nil {
			return err
		}
		if err := runGit("submodule init", "-C", root, "submodule", "init"); err != nil {
			return err
		}
		if err := runGit("submodule update", "-C", root, "submodule", "update"); err != nil {
			return err
		}
	} else {
		if err := runGit("pull", "-C", root, "pull"); err != nil {
			return err
		}
		if err := runGit("submodule update", "-C", root, "submodule", "update"); err != nil {
			return err
		}
	}
	return nil
}

func runGit(action string, args ...string) error {
	log.Infof("Found root directory - %sing updates!", action)
	log.Debugf("Commands %#v!", args)
	cmd := exec.Command("git", args...)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Could not start %sing: %w\n", action, err)
	}
	log.Infof("Waiting for git %s to finish...", action)
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Could not %s: %w\n", action, err)
	}

	log.Infof("Git %s finished!", action)
	return nil
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

	if err := getContent(); err != nil {
		panic(err)
	}

	root := getRoot()
	log.WithField("Root", root).Info("Our root directory")

	isReception, err := getDarkmode()
	if err != nil {
		panic(err)
	}

	// We'll parse and store the responses ahead of time.
	resps, err := pages.Load(isReception, root)
	if err != nil {
		log.Fatalf("pages.Load: unexpected error: %s", err)
	}
	log.WithField("Resps", resps).Debug("The parsed responses")
	responses = Atomic{Resps: resps}

	log.Info("Starting server.")
	log.Info("Listening on port: ", port)

	updateJumpFile(root)

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
			log.Warningln("notify.Watch:", err)
		}
		defer notify.Stop(events)

		go func() {
			for range events {
				if err := reloadContent(); err != nil {
					log.Warningln("Could not reload content: ", err)
				}
			}
		}()
	}

	// Listen on port and serve with our handler.
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func updateJumpFile(root string) {
	// If jumpfile exists.
	if _, err := os.Stat(root + "/jumpfile.json"); err == nil {
		buf, err := os.ReadFile(root + "/jumpfile.json")
		if err != nil {
			log.Warningln("jumpfile readfile: unexpected error:", err)
		}
		var j map[string]interface{}
		err = json.Unmarshal(buf, &j)
		if err != nil {
			log.Warningln("jumpfile unmarshal: unexpected error:", err)
		}
		log.Debugln(j)
		jumpfile = j
	} else {
		log.Infoln("No jumpfile found")
	}
}

// handler parses and serves responses to our file queries.
func handler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Access-Control-Allow-Origin", "*")
	res.Header().Add("Access-Control-Allow-Methods", "*")

	if v, ok := jumpfile[filepath.Clean(req.URL.Path)]; ok {
		newURL := v.(string)
		http.Redirect(res, req, newURL, http.StatusSeeOther)
		log.Infoln("Redirect: " + newURL)
		return
	}
	if req.URL.Path == "/fuzzyfile" {
		log.Info("Fuzzyfile")
		responses.Lock()
		buf, err := json.Marshal(fuzz.NewFile(responses.Resps))
		responses.Unlock()
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
	// NOTE: we're not checking the authenticity of any of these webhooks, but
	// since we're not getting any interesting data from them but instead
	// pulling that from either github or darkmode (using https so we can trust
	// that) the worst someone could do is a DOS.
	if req.Header.Get("X-Github-Event") == "push" {
		log.Infoln("Push hook")
		if err := getContent(); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := reloadContent(); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	if req.Header.Get("X-Darkmode-Event") == "updated" {
		log.Infoln("Darkmode hook")
		if err := reloadContent(); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	// Requested URL. We extract the path.
	query := req.URL.Path
	log.WithField("query", query).Info("Received query")

	clean := filepath.Clean(query)
	log.WithField("clean", clean).Info("Sanitized path")
	log.Println(rootDir(clean))
	responses.Lock()
	r, ok := responses.Resps[clean]
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
			responses.Resps[slug].Expanded,
			responses.Resps[slug].Sort,
		)
		// }
	}
	responses.Unlock()
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

func reloadContent() error {
	isReception, err := getDarkmode()
	if err != nil {
		return fmt.Errorf("Could not get darkmode status: %w", err)
	}
	resps, err := pages.Load(isReception, getRoot())
	if err != nil {
		return fmt.Errorf("Could not load pages: %w", err)
	}
	responses.Lock()
	responses.Resps = resps
	responses.Unlock()

	updateJumpFile(getRoot())

	return nil
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

var darkmode struct {
	mu     sync.Mutex
	result bool
}

func getDarkmode() (bool, error) {
	darkmode.mu.Lock()
	defer darkmode.mu.Unlock()

	url := getEnv("DARKMODE_URL")
	if url == "true" {
		darkmode.result = true
		return true, nil
	}
	if url == "false" {
		darkmode.result = false
		return false, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return true, err
	}
	if err := json.NewDecoder(res.Body).Decode(&darkmode.result); err != nil {
		return true, err
	}
	log.Info("Darkmode status: ", darkmode.result)

	return darkmode.result, nil
}
