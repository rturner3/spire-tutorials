#!/bin/bash
set -euo pipefail

spire_server_cmd="/bin/sh -c '/opt/spire/bin/spire-server $@ -socketPath /run/spire-server/private/api.sock'"
spire_server_docker_exec_cmd='sudo docker exec $(sudo docker ps | grep spire-server | awk '"'"'{print $1}'"'"')'" $spire_server_cmd"
pushd deployments/ubuntu1 >/dev/null
vagrant ssh -c "$spire_server_docker_exec_cmd"
popd >/dev/null
