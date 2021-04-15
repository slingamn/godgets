.PHONY: all test gofmt

all: test

test:
	go test . && go vet .
	./.check-gofmt.sh

gofmt:
	./.check-gofmt.sh --fix
