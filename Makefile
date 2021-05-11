.SILENT: sim bld tidy now

bld:
	cd cmd && go build main.go

trades: bld
	cmd/main -domain=trades -username="${u}"

sim: bld
	cmd/main -domain=sim -username="${u}"

now: bld
	cmd/main -domain=now -username="${u}"

tidy: bld
	cmd/main -domain=tidy -username="${u}"

user: bld
	cmd/main -domain=user -username="${u}"