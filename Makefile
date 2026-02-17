build:
	go build ride-sharing-system -o bin/fs

run: build
	./bin/fs

test:
	go test ./... -v