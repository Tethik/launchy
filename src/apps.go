package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type DesktopApp struct {
	Name         string
	Icon         string // https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html
	ShortcutPath string
	Exec         string
	Score        int
	// TODO Add more fields here, some Names are unfortunately not very descriptive
}

func execFieldToCmd(s string) (*exec.Cmd, error) {
	// Needs some work/testing.
	// Spec: https://specifications.freedesktop.org/desktop-entry-spec/desktop-entry-spec-latest.html#exec-variables
	args := []string{}
	b := strings.Builder{}
	quoted := false
	esc := false
	for _, r := range s {
		if r == ' ' && !quoted {
			arg := b.String()
			arg = strings.Trim(arg, " 	\"")
			if len(arg) > 0 {
				args = append(args, arg)
			}
			b.Reset()
		} else if r == '"' && !esc {
			quoted = !quoted
		} else if r == '\\' {
			esc = true
		} else {
			esc = false
		}

		_, err := b.WriteRune(r)
		panicIf(err)
	}

	if b.Len() > 0 {
		arg := b.String()
		arg = strings.Trim(arg, " 	\"")
		if len(arg) > 0 {
			args = append(args, arg)
		}
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("exec is empty: %s", s)
	}

	log.Info(args)
	cmd := exec.Command(args[0], args[1:]...)
	return cmd, nil
}

func (app *DesktopApp) Start() error {
	log.Infof("ShortcutPath: %s", app.ShortcutPath)
	log.Infof("Cmd: %s", app.Exec)

	cmd, err := execFieldToCmd(app.Exec)
	if err != nil {
		return err
	}
	err = cmd.Start()
	return err
}

func formatExecString(exec string) string {
	// https://specifications.freedesktop.org/desktop-entry-spec/desktop-entry-spec-latest.html#exec-variables
	// Exec allows for passing certain arguments. Some applications have these, so we need to either fill
	// in the correct value or remove it. Here are all the possible args:
	// "%f"	A single file name (including the path), even if multiple files are selected. The system reading the desktop entry should recognize that the program in question cannot handle multiple file arguments, and it should should probably spawn and execute multiple copies of a program for each selected file if the program is not able to handle additional file arguments. If files are not on the local file system (i.e. are on HTTP or FTP locations), the files will be copied to the local file system and %f will be expanded to point at the temporary file. Used for programs that do not understand the URL syntax.
	// "%F"	A list of files. Use for apps that can open several local files at once. Each file is passed as a separate argument to the executable program.
	// "%u"	A single URL. Local files may either be passed as file: URLs or as file path.
	// "%U"	A list of URLs. Each URL is passed as a separate argument to the executable program. Local files may either be passed as file: URLs or as file path.
	// "%i"	The Icon key of the desktop entry expanded as two arguments, first --icon and then the value of the Icon key. Should not expand to any arguments if the Icon key is empty or missing.
	// "%c"	The translated name of the application as listed in the appropriate Name key in the desktop entry.
	// "%k"	The location of the desktop file as either a URI (if for example gotten from the vfolder system) or a local filename or empty if no location is known.
	// "%d", "%D", "%n", "%N", "%v", "%m" Are all deprecated

	toRemove := []string{"%f", "%F", "%u", "%U", "%i", "%c", "%k", "%d", "%D", "%n", "%N", "%v", "%m"}
	for _, r := range toRemove {
		exec = strings.ReplaceAll(exec, r, "")
	}

	// TODO: implement %i, %c, %k (tho I think unlikely these are used often)
	return exec
}

// Use channel later..
func parseFile(f string) *DesktopApp {
	file, err := os.Open(f)
	if err != nil {
		log.Warnf("Could not open %s for parsing:\n %s", f, err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var app DesktopApp
	app.ShortcutPath = f
	seen := map[string]bool{}
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == '[' || line[0] == ']') {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])

		// Hack to only parse first entry. A desktop file can contain multiple
		// entries (see firefox), so the proper solution would be to parse this better.
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = true

		value := parts[1]
		switch key {
		case "Name":
			app.Name = value
			continue
		case "Icon":
			app.Icon = value
			continue
		case "Exec":
			app.Exec = formatExecString(value)
			continue
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return &app
}

func enumerateDirForApps(dir string) (l []*DesktopApp) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Warnf("Listing %s failed:\n %s", dir, err)
		return
	}
	for _, f := range files {
		p := path.Join(dir, f.Name())
		finfo, err := os.Stat(p)
		if finfo.IsDir() {
			continue
		}
		if err != nil {
			log.Warnf("Errd on stat %s:\n %s", p, err)
			continue
		}

		a := parseFile(p)
		if a != nil {
			l = append(l, a)
		}
	}
	return
}

func applicationDirs() []string {
	home, err := os.UserHomeDir()
	// i dont wanna deal
	panicIf(err)

	paths := []string{
		path.Join(home, ".local/share/applications/"),
		"/usr/share/applications/",
	}

	envs := environMap()

	xdd, ok := envs["XDG_DATA_DIRS"]
	if ok {
		dds := strings.Split(xdd, ":")
		for _, d := range dds {
			paths = append(paths, path.Join(d, "applications/"))
		}
	}

	return paths
}

type Searcher struct {
	data   []*DesktopApp
	scores map[string]int
}

func loadScores() map[string]int {
	scores := map[string]int{}
	home, err := os.UserHomeDir()
	panicIf(err)
	f := path.Join(home, ".local/share/launchy/scores.json")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		log.Warnf("Could not open %s to fetch scores:\n %s", f, err)
		return scores
	}
	err = json.Unmarshal(b, &scores)
	warnIf(err)
	return scores
}

func (s *Searcher) saveScores() {
	home, err := os.UserHomeDir()
	panicIf(err)
	d := path.Join(home, ".local/share/launchy/")
	f := path.Join(d, "scores.json")
	b, err := json.Marshal(s.scores)
	panicIf(err)
	os.MkdirAll(d, os.ModePerm)
	err = ioutil.WriteFile(f, b, os.FileMode(int(0640)))
	warnIf(err)
}

func (s *Searcher) AddScore(app string) {
	n := s.scores[app]
	s.scores[app] = n + 1
	s.saveScores()
}

func SearcherNew() *Searcher {
	// TODO: might be worth caching / indexing here somehow.
	var all []*DesktopApp
	paths := applicationDirs()
	log.Infof("Application dirs:\n %s", paths)

	scores := loadScores()

	seen := map[string]bool{}
	for _, p := range paths {
		apps := enumerateDirForApps(p)
		for _, a := range apps {
			// Lazy way to filter out duplicates
			if _, ok := seen[a.Name]; !ok {
				a.Score = scores[a.Name]
				all = append(all, a)
				seen[a.Name] = true
			}
		}
	}

	searcher := Searcher{all, scores}
	return &searcher
}

func (s *Searcher) SearchApps(text string) (result []*DesktopApp) {
	for _, a := range s.data {
		// Lazy solution for now.
		if strings.Contains(strings.ToLower(a.Name), strings.ToLower(text)) {
			result = append(result, a)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		// higher score goes first
		return result[i].Score > result[j].Score
	})
	return
}
