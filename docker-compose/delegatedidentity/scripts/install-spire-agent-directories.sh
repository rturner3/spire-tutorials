#!/usr/bin/bash

install -o root -g root -m 755 -d /etc/spire-agent /run/spire-agent/.data
install -o root -g root -m 600 -D ./configs/spire-agent/* /etc/spire-agent
