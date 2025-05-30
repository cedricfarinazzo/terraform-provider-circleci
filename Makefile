.PHONY: build test clean

build:
	go build -o terraform-provider-circleci

test:
	go test ./...

clean:
	rm -f terraform-provider-circleci

check: test
	go fmt ./...
	go vet ./...
