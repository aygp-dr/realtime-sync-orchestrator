.PHONY: build run test clean

build:
	go build -o bin/realtime-sync-orchestrator .

run: build
	./bin/realtime-sync-orchestrator

test:
	go test ./...

clean:
	rm -rf bin/
