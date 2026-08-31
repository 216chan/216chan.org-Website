.PHONY: db-up db-down backend frontend install

db-up:
	docker compose up -d db

db-down:
	docker compose down

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

install:
	cd frontend && npm install
	cd backend && go mod tidy
