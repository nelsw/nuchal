.SILENT: up down bld sim trade it

bld:
	PATH=${PATH}:/Users/${USER}/go/bin && go install

it:
	GOOS=linux GOARCH=amd64 && go build -o build/nuchal main.go

sim: it up
	build/nuchal sim && open http://localhost:8090

report: it
	build/nuchal report

trade: it
	build/nuchal trade

up:
	docker compose -p nuchal -f build/docker-compose.yml up -d

down:
	docker compose -f build/docker-compose.yml down