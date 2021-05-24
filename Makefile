.SILENT: up down bld sim trade it

up:
	docker compose -p nuchal -f build/docker-compose.yml up -d

down:
	docker compose -f build/docker-compose.yml down