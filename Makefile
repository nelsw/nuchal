.SILENT: up down it

it:
	GOOS=linux GOARCH=amd64 && go build -o nuchal main.go

up:
	docker compose -f docker/docker-compose.yml up --build --force-recreate --remove-orphans

down:
	docker compose -f docker/docker-compose.yml down