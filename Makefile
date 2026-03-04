build:
	go build -o issues .

install: build
	cp issues /usr/local/bin/issues

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f issues

.PHONY: build install test lint clean
