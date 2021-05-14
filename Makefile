.SILENT: up down

up:
	docker compose -f docker/docker-compose.yml up --build --force-recreate -d

down:
	docker compose -f docker/docker-compose.yml down