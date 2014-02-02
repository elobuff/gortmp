default: fmt test

fmt:
	go fmt

test:
	go test -cover
