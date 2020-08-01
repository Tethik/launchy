package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/prometheus/common/log"
)

type DesktopApp struct {
	Name string
	Icon string // https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html
	Exec string
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

		value := strings.TrimSpace(parts[1])
		switch key {
		case "Name":
			app.Name = value
			break
		case "Icon":
			app.Icon = value
			break
		case "Exec":
			app.Exec = value
			break
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

func SearchApps(text string) (result []*DesktopApp) {
	// TODO: add any sort of caching / indexing here.
	// not great to open/read all these files. For now, it actually works pretty well.
	// TODO: remove duplicate entries (same shortcut exists for multiple apps)
	var all []*DesktopApp
	paths := applicationDirs()
	log.Infof("Application dirs:\n %s", paths)
	for _, p := range paths {
		all = append(all, enumerateDirForApps(p)...)
	}

	for _, a := range all {
		// Lazy solution for now.
		if strings.Contains(strings.ToLower(a.Name), strings.ToLower(text)) {
			result = append(result, a)
		}
	}
	return
}
