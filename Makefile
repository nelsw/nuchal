.SILENT: sim tgt hst bld it

bld:
	cd cmd && go build main.go

cht: bld
	echo "\n...build ${s} charts"
	cmd/main cht ${s} "${u}"
	echo "\n...built ${s} charts"

hst: bld
	echo "\n...build ${s} history"
	cmd/main hst ${s} "${u}"
	echo "\n...built ${s} history"

it: bld
	echo "\n...run ${s} trades for ${u}"
	cmd/main trade ${s} "${u}"

sim: bld
	echo "\n...run ${s} simulation for ${u}\n"
	cmd/main sim ${s} "${u}"
	echo "\n...ran ${s} simulation for ${u}\n"
