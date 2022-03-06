build:
	go build -o promux *.go

install:
	cp promux /usr/bin/promux

clear:
	rm promux
