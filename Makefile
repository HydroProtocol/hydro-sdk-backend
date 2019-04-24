test:
	go test ./... --count=1 --cover

clean:
	go clean

.PHONY: test api ws watcher engine launcher
