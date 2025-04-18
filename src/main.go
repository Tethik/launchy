package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	// TODO: fix this dependency. It's a nice log tho
	log "github.com/sirupsen/logrus"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type Application struct {
	win           *gtk.Window
	resultList    *gtk.ListBox
	searchbar     *gtk.Entry
	currentResult []*DesktopApp
	css           *gtk.CssProvider
	searcher      *Searcher
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

	app.currentResult = app.searcher.SearchApps(searchText)

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
	// TODO: Some messyness here... need to look up how Exec is supposed to work. Figure out if
	// bash should be used. Maybe use $SHELL ?
	err := app.currentResult[i].Start()
	if err != nil {
		log.Warnf("%s", err)
		// TODO: give feedback to user
	} else {
		// So long as the process exits cleanly, it seems the children live on.
		app.searcher.AddScore(app.currentResult[i].Name)
		gtk.MainQuit()
	}
}

func (app *Application) addSearchResultItem(item *DesktopApp) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	panicIf(err)

	ctx, _ := row.GetStyleContext()
	ctx.AddClass("result_row")

	var icon *gtk.Image
	if len(item.Icon) > 0 && item.Icon[0] == '/' {
		pixbuf, pErr := gdk.PixbufNewFromFileAtSize(item.Icon, 64, 64)
		err = pErr
		if pErr == nil {
			icon, err = gtk.ImageNewFromPixbuf(pixbuf)
		} else {
			log.Warnf("Failed to load pixbuf from img: %s, err:\n", item.Icon, err)
		}
	} else {
		icon, err = gtk.ImageNewFromIconName(item.Icon, gtk.ICON_SIZE_DIALOG)
		// Icons dont seem to play nice
		// pixbuf := icon.GetPixbuf()

		// fmt.Println(item.Icon, icon.GetAllocatedWidth(), pixbuf.GetWidth())
		// if pixbuf.GetWidth() > 64 {
		// pixbuf, pErr := icon.GetPixbuf().ScaleSimple(64, 64, gdk.INTERP_BILINEAR)
		// err = pErr
		// if pErr == nil {
		// 	icon, err = gtk.ImageNewFromPixbuf(pixbuf)
		// }
		// }
	}

	if err != nil {
		log.Warnf("Failed to load icon for app: %s", item.Icon)
		icon, err = gtk.ImageNewFromIconName("model", gtk.ICON_SIZE_DIALOG)
		if err != nil {
			log.Warnf("Fallback icon failed to load too: %s", err)
		}
	}

	if icon != nil {
		ctx, _ = icon.GetStyleContext()
		ctx.AddClass("app_icon")
		row.Add(icon)
	}

	label, err := gtk.LabelNew(item.Name)
	panicIf(err)

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
	win.SetDecorated(false)                      // removes all borders and window title
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG) // Tells the VM that this window is a dialog, allowing it to float by default in i3.
	win.SetGravity(gdk.GDK_GRAVITY_CENTER)       // sets origo in center
	win.Move(0, 0)                               // centers the app
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

	searcher := SearcherNew()

	app := &Application{
		win,
		list,
		entry,
		[]*DesktopApp{},
		css,
		searcher,
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
		// log.Info("Key pressed: %s", keyEvent.KeyVal())

		// fmt.Println(keyEvent.State(), gdk.GDK_CONTROL_MASK)
		// fmt.Println(keyEvent.KeyVal(), gdk.KEY_q)

		if keyEvent.KeyVal() == gdk.KEY_Return {
			app.handleLaunch()
			return
		}
		if keyEvent.KeyVal() == gdk.KEY_Escape || (keyEvent.KeyVal() == gdk.KEY_q && keyEvent.State()&gdk.CONTROL_MASK > uint(0)) {
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

		// Ensures searchbar gets all input
		app.searchbar.GrabFocusWithoutSelecting()
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

const lockFile = "/tmp/launchy.lock"

func main() {
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err == nil {
		defer os.Remove(lockFile)
		defer file.Close()

		// Try to lock the file
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			log.Info("Another instance of Launchy is already running.")
			return
		}
		defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	} else {
		log.Warnf("Failed to create/open lock file: %s", err)
		return
	}

	fmt.Println("Starting Launchy...")

	app := NewApplication()
	app.Main()

	fmt.Println("Closing Launchy...")
}
