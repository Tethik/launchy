# README

A very basic application launcher for Linux. It scans desktop entries for applications, and that's all. No background
process required.

## Dependencies
Uses Go 1.14 (and [gotk3](https://github.com/gotk3/gotk3)), but may be compatible with other versions.

Other library dependencies for compiling on Ubuntu:
```
libgtk-3-dev
libglib2.0-dev
libgdk-pixbuf2.0-dev
```

## Installing

```sh
make # note: fetching gotk3 for the first time may take a while.
sudo make install
```

If you want to add some power/system shortcuts, you can use the ones
in the `powershortcuts/` folder. You probably want to configure them though, 
since they are made with icons/commands that work for my own i3 setup.

`make shortcuts` will copy them over to the `~/.local/share/applications` directory.

### i3 configuration
Add the following to your i3 config (`~/.config/i3/config`) to launch
the application via $mod+d (typicall Win+D).

```
bindsym $mod+d exec launchy
```