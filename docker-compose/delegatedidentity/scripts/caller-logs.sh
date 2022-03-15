#!/bin/bash
set -euo pipefail

for i in 1 2 3; do
	echo "BEGIN CALLER$i ==================================="
	pushd deployments/ubuntu2 >/dev/null
	find_id_cmd="sudo docker ps | grep caller$i | awk '{print "'$1'"}'"
	cmd='sudo docker logs $('"$find_id_cmd"')'
	vagrant ssh -c "$cmd" 2>/dev/null
	echo "END CALLER$i ==================================="
	popd >/dev/null
done
