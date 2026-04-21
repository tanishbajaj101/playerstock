.PHONY: up down backend frontend dev

d:
	docker compose up -d

down:
	docker compose down

b:
	cd backend && go run ./cmd/server

f:
	cd frontend && npm run dev

dev:
	docker compose up -d
	cd backend && go run ./cmd/server &
	cd frontend && npm run dev
