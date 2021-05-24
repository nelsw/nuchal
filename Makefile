.SILENT: up down bld sim trade it

it: bld
	cd build && go install && export PATH=$PATH:/Users/${USER}/go/bin

bld:
	GOOS=linux GOARCH=amd64 && go build -o build/nuchal main.go

sim: bld up
	MODE=dev build/nuchal sim && open http://localhost:8090

report: bld
	build/nuchal report

trade: bld
	build/nuchal trade

up:
	docker compose -p nuchal -f build/docker-compose.yml up -d

down:
	docker compose -f build/docker-compose.yml down