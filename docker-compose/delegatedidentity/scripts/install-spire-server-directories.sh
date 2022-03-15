#!/usr/bin/bash

install -o root -g root -m 755 -d /etc/spire-server /run/spire-server/.data /var/lib/spire-server
install -o root -g root -m 600 -D ./configs/spire-server/* /etc/spire-server
