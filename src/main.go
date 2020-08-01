package main

import (
	"os/exec"
	"strings"

	// TODO: fix this dependency. It's a nice log tho
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
	css           *gtk.CssProvider
}

func (app *Application) handleSearch() {
	searchText, err := app.searchbar.GetText()
	panicIf(err)

	searchText = strings.TrimSpace(searchText)

	// Clear previous results
	for app.resultList.GetChildren().Length() > 0 {
		n := app.resultList.GetChildren().Data().(gtk.IWidget)
		app.resultList.Remove(n)
	}

	if len(searchText) == 0 {
		return
	}

	app.currentResult = SearchApps(searchText)

	for _, item := range app.currentResult {
		app.addSearchResultItem(item)
	}

	if len(app.currentResult) > 0 {
		app.resultList.SelectRow(app.resultList.GetRowAtIndex(0))
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
	// TODO: Some messyness here... need to look up how Exec is supposed to work.
	// lots of app take extra arguments, e.g. files or urls.
	cmd := exec.Command("bash", "-c", app.currentResult[i].Exec)
	err := cmd.Start()
	if err != nil {
		log.Warnf("%s", err)
	}
}

func (app *Application) addSearchResultItem(item *DesktopApp) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	panicIf(err)

	ctx, _ := row.GetStyleContext()
	ctx.AddClass("result_row")

	var icon *gtk.Image
	if len(item.Icon) > 0 && item.Icon[0] == '/' {
		icon, err = gtk.ImageNewFromFile(item.Icon)
	} else {
		icon, err = gtk.ImageNewFromIconName(item.Icon, gtk.ICON_SIZE_LARGE_TOOLBAR)
	}

	if err != nil {
		log.Warnf("Failed to load icon for app: %s", item.Icon)
	}

	ctx, _ = icon.GetStyleContext()
	ctx.AddClass("app_icon")

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

	css, err := gtk.CssProviderNew()
	panicIf(err)

	// err = css.LoadFromPath("style.css")
	err = css.LoadFromData(stylesheet)
	panicIf(err)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	gtk.AddProviderForScreen(win.GetScreen(), css, 0)

	win.SetTitle("Launchy")
	win.SetDecorated(false) // removes all borders and window title
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	ctx, _ := win.GetStyleContext()
	ctx.AddClass("window")

	layoutList, err := gtk.ListBoxNew()
	layoutList.SetSelectionMode(gtk.SELECTION_NONE)
	panicIf(err)

	hadj, err := gtk.AdjustmentNew(0, 0, 640, 50, 640, 640)
	panicIf(err)
	vadj, err := gtk.AdjustmentNew(0, 0, 640, 50, 640, 640)
	panicIf(err)
	scroll, err := gtk.ScrolledWindowNew(hadj, vadj)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	scroll.SetSizeRequest(-1, 480)
	panicIf(err)

	list, err := gtk.ListBoxNew()
	panicIf(err)
	list.SetSelectionMode(gtk.SELECTION_SINGLE)
	ctx, _ = list.GetStyleContext()
	ctx.AddClass("result_list")

	entry, _ := gtk.EntryNew()
	ctx, _ = entry.GetStyleContext()
	ctx.AddClass("search_bar")

	layoutList.Add(entry)
	scroll.Add(list)
	layoutList.Add(scroll)

	win.Add(layoutList)

	app := &Application{
		win,
		list,
		entry,
		[]*DesktopApp{},
		css,
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

		// fmt.Println(keyEvent.State(), gdk.GDK_CONTROL_MASK)
		// fmt.Println(keyEvent.KeyVal(), gdk.KEY_q)

		if keyEvent.KeyVal() == gdk.KEY_Return {
			app.handleLaunch()
			return
		}
		if keyEvent.KeyVal() == gdk.KEY_Escape || (keyEvent.KeyVal() == gdk.KEY_q && keyEvent.State()&gdk.GDK_CONTROL_MASK > uint(0)) {
			gtk.MainQuit()
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
			i = i - 1
			if i < 0 {
				i = l - 1
			}
		}
		// TODO: should scroll too.
		app.resultList.SelectRow(app.resultList.GetRowAtIndex(i))
	})

	entry.GrabFocusWithoutSelecting()

	// Set the default window size.
	win.SetDefaultSize(800, 250)

	return app
}

func (app *Application) Main() {
	app.win.ShowAll()
	gtk.Main()
}

func main() {
	app := NewApplication()
	app.Main()
}
