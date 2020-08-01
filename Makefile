all: launchy

launchy: src/*.go
	go build -o launchy ./src/

debug: launchy
	env GTK_DEBUG=interactive ./launchy

clean:
	rm launchy

install: launchy
	cp launchy /usr/local/bin/