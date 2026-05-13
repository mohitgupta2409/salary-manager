.PHONY: build run test seed clean

build:
	go build -o bin/server ./main.go

run:
	go run ./main.go

test:
	go test ./... -v -count=1

seed:
	go run ./cmd/seed/main.go

clean:
	rm -f bin/server
	rm -f salary-manager.db
