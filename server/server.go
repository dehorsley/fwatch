package server

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"sync"
	"text/template"
	"time"

	"fwatch/db"

	"github.com/BurntSushi/toml"
	"github.com/radovskyb/watcher"
)

// Server stores the main state of the server
type Server struct {
	SettleTime *duration             `toml:"settle"`
	PollPeriod *duration             `toml:"poll"`
	DBPath     string                `toml:"database"`
	Groups     map[string]*FileGroup `toml:"files"`

	db      *db.Database
	done    chan struct{}
	uploads sync.WaitGroup
}

// New creates a new server from a toml config. Will return an error if the
// config is incorrect.
func New(config io.Reader) (*Server, error) {
	s := &Server{done: make(chan struct{})}
	_, err := toml.DecodeReader(config, s)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %s", err)
	}

	if s.DBPath == "" {
		return nil, fmt.Errorf("'database' not set in config")
	}

	if s.Groups == nil {
		return nil, fmt.Errorf("no groups set in config")
	}

	s.db, err = db.Open(s.DBPath)

	if err != nil {
		return nil, fmt.Errorf("opening database: %s", err)
	}

	if s.PollPeriod == nil {
		s.PollPeriod = &duration{5 * time.Minute}
	}

	if s.SettleTime == nil {
		s.SettleTime = &duration{20 * time.Minute}
	}

	for groupName, g := range s.Groups {
		if g.PollPeriod == nil {
			g.PollPeriod = s.PollPeriod
		}

		if g.SettleTime == nil {
			g.SettleTime = s.SettleTime
		}

		err := g.setupWatcher()
		if err != nil {
			return nil, fmt.Errorf("setting up group watcher %s: %s", groupName, err)
		}
	}
	return s, nil
}

// Start runs the server in the foreground
func (s *Server) Start() error {
	var wg sync.WaitGroup

	errs := make(chan error)

	go func() {
		for _, g := range s.Groups {
			for path, info := range g.w.WatchedFiles() {
				select {
				case <-s.done:
					return
				default:
					if s.db.Last(path).Before(info.ModTime()) {
						s.upload(path)
					}
				}
			}
		}
	}()

	for _, g := range s.Groups {
		wg.Add(1)
		go func(g *FileGroup) {
			s.handleGroup(g)
			wg.Done()
		}(g)

		wg.Add(1)
		go func(g *FileGroup) {
			err := g.w.Start(g.PollPeriod.Duration)
			if err != nil {
				errs <- err
			}
			wg.Done()
		}(g)
	}

	var err error
	select {
	case <-s.done:
	case err = <-errs:
	}

	for _, g := range s.Groups {
		g.w.Close()
	}
	wg.Wait()

	uploadsDone := make(chan struct{})
	go func() {
		s.uploads.Wait()
		close(uploadsDone)
	}()

	// Let the client know why we're paused
	select {
	case <-time.After(200 * time.Millisecond):
		log.Println("waiting for uploads to complete...")
	case <-uploadsDone:
	}
	<-uploadsDone

	s.db.Close()
	return err
}

// Ignore ignores changes to the file given in path
func (s *Server) Ignore(paths ...string) {
	s.db.Update(time.Now(), paths...)
}

// Reset removes upload record of path
func (s *Server) Reset(path ...string) {
	s.db.Remove(path...)
}

// ListUploads returns a list of all upload times and
func (s *Server) ListUploads() map[string]time.Time {
	m := make(map[string]time.Time)
	s.db.ForEach(func(path string, t time.Time) error {
		m[path] = t
		return nil
	})
	return m
}

func (s *Server) upload(path string) {
	s.uploads.Add(1)
	log.Printf("uploading %s...\n", filepath.Base(path))
	time.Sleep(5 * time.Second)
	log.Printf("uploading %s... done\n", filepath.Base(path))
	s.db.Update(time.Now(), path)
	s.uploads.Done()

}

func (s *Server) deferredUpload(path string) func() {
	return func() {
		s.upload(path)
	}
}

func (s *Server) handleGroup(g *FileGroup) {
	jobs := make(map[string]*time.Timer)

Loop:
	for {
		select {
		case e := <-g.w.Event:
			if e.IsDir() {
				continue
			}

			switch e.Op {
			case watcher.Remove, watcher.Rename:
				if t, ok := jobs[e.Path]; ok {
					t.Stop()
					delete(jobs, e.Path)
				}
			default:
				if t, ok := jobs[e.Path]; ok {
					t.Reset(g.SettleTime.Duration)
					continue
				}
				jobs[e.Path] = time.AfterFunc(g.SettleTime.Duration, s.deferredUpload(e.Path))
			}

		case err := <-g.w.Error:
			log.Fatalln(err)
		case <-g.w.Closed:
			break Loop
		}
	}

	for _, timer := range jobs {
		timer.Stop()
	}
}

// Stop gracefully stops the server. Shutdown has completed after Start call exits
func (s *Server) Stop() {
	close(s.done)
}

func (g *FileGroup) setupWatcher() (err error) {
	g.w = watcher.New()

	if g.FilenameMatch != nil {
		g.w.AddFilterHook(watcher.RegexFilterHook(g.FilenameMatch.Regexp, false))
	}

	for _, p := range g.Paths {
		if g.Recursive {
			err = g.w.AddRecursive(p)
		} else {
			err = g.w.Add(p)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// FileGroup represents a groups of files that will be polled and uploaded together
type FileGroup struct {
	FilenameMatch   *re          `toml:"filename_match"`
	FileContent     string       `toml:"file_type"`
	FileTypeContent string       `toml:"file_type_content"`
	SettleTime      *duration    `toml:"settle_time"`
	PollPeriod      *duration    `toml:"poll_period"`
	EmailTemplate   textTemplate `toml:"email_template"`
	Email           []string
	Paths           []string
	Recursive       bool

	w *watcher.Watcher
}

// List returns a list of paths watched by the file group
func (g *FileGroup) List() (paths []string) {
	f := g.w.WatchedFiles()

	for k := range f {
		paths = append(paths, k)
	}
	// sort.Strings(paths)
	return paths
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type re struct {
	*regexp.Regexp
}

func (r *re) UnmarshalText(text []byte) error {
	var err error
	r.Regexp, err = regexp.Compile(string(text))
	return err
}

type textTemplate struct {
	*template.Template
}

func (t *textTemplate) UnmarshalText(text []byte) (err error) {
	t.Template, err = template.New("email").Parse(string(text))
	return err
}
