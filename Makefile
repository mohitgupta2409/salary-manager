.PHONY: all backend frontend test seed clean

all: backend frontend

## Backend
backend:
	cd backend && go build -o bin/server ./main.go

backend-run:
	cd backend && go run ./main.go

backend-test:
	cd backend && go test ./... -v -count=1

## Frontend
frontend:
	cd frontend && npm install && npm run build

frontend-dev:
	cd frontend && npm run dev

## Seed database with 10K employees
seed:
	cd backend && go run ./cmd/seed/main.go

## Run both (backend + frontend dev)
dev:
	@echo "Start backend: make backend-run"
	@echo "Start frontend: make frontend-dev"

## Test everything
test: backend-test

## Clean build artifacts
clean:
	rm -f backend/bin/server
	rm -rf frontend/dist
	rm -f backend/salary-manager.db
