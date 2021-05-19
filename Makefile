.SILENT: up down it sim trade acc

it:
	GOOS=linux GOARCH=amd64 && go build -o nuchal main.go

sim: it up
	./nuchal sim

account: it
	./nuchal account

holds: it
	./nuchal account --force-holds

trade: it
	./nuchal trade

up:
	docker compose -f build/docker-compose.yml up --build --force-recreate --remove-orphans -d

down:
	docker compose -f build/docker-compose.yml down