.SILENT: sim tgt hst bld it

bld:
	cd cmd && go build main.go

it: bld
	cmd/main -domain=trade -symbol=${s} -username="${u}"

trades: bld
	cmd/main -domain=trades -symbols=${s} -username="${u}"

sim: bld
	cmd/main -domain=sim -symbol=${s} -username="${u}"

user: bld
	cmd/main -domain=user -username="${u}" -key=${k} -pass=${p} -secret=${s}