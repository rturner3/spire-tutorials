#!/usr/bin/bash

apt-get update
apt-get upgrade -y

apt-get install -y \
    build-essential \
    ca-certificates \
    curl \
    dkms \
    git \
    gnupg \
    linux-headers-$(uname -r) \
    lsb-release 
