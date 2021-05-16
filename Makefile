.SILENT: up down

up:
	docker compose -f docker/docker-compose.yml up --build --force-recreate --remove-orphans

down:
	docker compose -f docker/docker-compose.yml down