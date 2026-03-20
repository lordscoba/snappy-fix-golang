dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

dev-bg:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build

dev-down:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down

dev-logs:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f api

prod:
	docker compose up --build -d

prod-down:
	docker compose down