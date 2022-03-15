#!/bin/bash
set -euo pipefail

cmd='sudo docker logs $(sudo docker ps | grep proxy | awk '"'"'{print $1}'"'"')'
pushd deployments/ubuntu2 >/dev/null
vagrant ssh -c "$cmd" 2>/dev/null
popd >/dev/null
