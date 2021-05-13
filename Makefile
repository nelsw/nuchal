.SILENT: sim bld tidy now

trades:
	cmd/main -domain=trade -name="${u}"

sim:
	cd docker/sim && docker-compose up --build

user:
	cd docker/user && docker-compose up --build