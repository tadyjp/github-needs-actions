.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o run
	zip -r run.zip run config
	mv run.zip build
	rm -f run
