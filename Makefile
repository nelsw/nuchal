.SILENT: it

it:
	cd docker && docker-compose up --build

trades:
	cmd/main -domain=trade -name="${u}"

sim:
	cd docker/sim && docker-compose up --build

user:
	cd docker/user && docker-compose up --build