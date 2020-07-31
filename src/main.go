package main

import (
	"fmt"
	"os/exec"

	"github.com/prometheus/common/log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func search(text string) error {
	return nil
}

type Application struct {
	win           *gtk.Window
	resultList    *gtk.ListBox
	searchbar     *gtk.Entry
	currentResult []*DesktopApp
}

func (app *Application) handleSearch() {
	searchText, err := app.searchbar.GetText()
	panicIf(err)
	fmt.Println(searchText)

	// Clear previous results
	for app.resultList.GetChildren().Length() > 0 {
		n := app.resultList.GetChildren().Data().(gtk.IWidget)
		app.resultList.Remove(n)
	}

	// fakeResult := []DesktopApp{
	// 	{Name: "Firefox"},
	// 	{Name: "Code"},
	// }

	app.currentResult = SearchApps(searchText)

	for _, item := range app.currentResult {
		app.addSearchResultItem(item)
	}

	app.resultList.ShowAll()
}

func (app *Application) handleLaunch() {
	log.Info("Got launch event")
	row := app.resultList.GetSelectedRow()
	if row == nil {
		log.Info("no selected row")
		return
	}

	i := row.GetIndex()
	// Some messyness here.. need to look up how Exec is supposed to work.
	cmd := exec.Command("bash", "-c", app.currentResult[i].Exec)
	err := cmd.Start()
	if err != nil {
		log.Warnf("%s", err)
	}
}

func (app *Application) addSearchResultItem(item *DesktopApp) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	panicIf(err)

	// Finding the icon is a bit trickier...
	// https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html
	icon, err := gtk.ImageNewFromFile(item.Icon)
	panicIf(err)

	label, err := gtk.LabelNew(item.Name)
	panicIf(err)

	row.Add(icon)
	row.Add(label)

	// Automatically inserts a GtkListBoxRow around the box
	app.resultList.Add(row)
}

func NewApplication() *Application {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	layoutList, err := gtk.ListBoxNew()
	panicIf(err)

	scroll, err := gtk.ScrolledWindowNew()
	panicIf(err)

	list, err := gtk.ListBoxNew()
	panicIf(err)
	list.SetSelectionMode(gtk.SELECTION_SINGLE)

	entry, err := gtk.EntryNew()

	layoutList.Add(entry)
	layoutList.Add(list)

	win.Add(layoutList)

	app := &Application{
		win,
		list,
		entry,
		[]*DesktopApp{},
	}

	// Bind events
	entry.Connect("changed", func() {
		app.handleSearch()
	})
	// entry.Connect("activate", func() {
	// 	app.handleLaunch()
	// })
	win.Connect("key-press-event", func(window *gtk.Window, event *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		log.Info("Key pressed: %s", keyEvent.KeyVal())
		if keyEvent.KeyVal() == gdk.KEY_Return {
			app.handleLaunch()
			return
		}

		i := app.resultList.GetSelectedRow().GetIndex()
		l := len(app.currentResult)
		if l == 0 {
			return
		}
		if keyEvent.KeyVal() == gdk.KEY_Down {
			i = (i + 1) % l
		}
		if keyEvent.KeyVal() == gdk.KEY_Up {
			i = (i - 1) % l
		}
		app.resultList.SelectRow(app.resultList.GetRowAtIndex(i))
	})

	// Set the default window size.
	win.SetDefaultSize(800, 250)

	return app
}

func (app *Application) Main() {
	// Recursively show all widgets contained in this window.
	app.win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}

func main() {

	app := NewApplication()

	app.Main()
}
