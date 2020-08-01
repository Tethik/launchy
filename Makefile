all: launchy

launchy:
	go build -o launchy ./src/

debug: launchy
	env GTK_DEBUG=interactive ./launchy

clean:
	rm launchy