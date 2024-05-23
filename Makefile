build:
	go build -o bin/app

run: build
	cd bin && ./app