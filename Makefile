.SILENT: up down build sim trade acc

build:
	GOOS=linux GOARCH=amd64 && go build -o build/nuchal main.go

sim: build up
	build/nuchal sim && open http://localhost:${SIM_PORT}

account: build
	build/nuchal account

trade: build
	build/nuchal trade

up:
	docker compose -p nuchal -f build/docker-compose.yml up -d

down:
	docker compose -f build/docker-compose.yml down