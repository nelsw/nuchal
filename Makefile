.SILENT: up down it sim trade acc

it:
	GOOS=linux GOARCH=amd64 && go build -o nuchal main.go

sim: it up
	./nuchal sim && open http://localhost:8089

account: it
	./nuchal account

holds: it
	./nuchal account --force-holds

trade: it
	./nuchal trade

up:
	docker compose up -d

down:
	docker compose down