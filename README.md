# README

A very basic application launcher for Linux. It scans desktop entries for applications, and that's all.

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