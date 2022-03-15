# SPIRE Delegated Identity APIs Proof of Concept
This proof of concept contains the following components:
- SPIRE Server v1.2.1
- SPIRE Agent v1.2.1
- A service named `callee` that exposes an API over a TLS-protected endpoint using an X.509-SVID from SPIRE Agent as the TLS server certificate, and logs the SPIFFE ID (URI SAN) of the client X.509 certificate
- A service named `proxy` that uses the SubscribeToX509SVIDs and SubscribeToX509Bundles APIs and forwards requests to the callee service endpoint, using the caller's X.509-SVID as the client certificate of the TLS connection
- Three services named `caller1`, `caller2`, and `caller3` that issue RPCs to the API exposed by the callee service through the proxy

The `proxy` exposes a Unix domain socket endpoint that `caller1`, `caller2`, and `caller3` use to send calls to `callee`.

# Run Proof of Concept
## Prerequisites
1. Install [Virtualbox](https://www.virtualbox.org/wiki/Downloads)
1. Install [Vagrant](https://www.vagrantup.com/downloads)
1. Install `vagrant-vbguest` Vagrant plugin
   ```bash
   $ vagrant plugin install vagrant-vbguest
   ```
1. Set up Ubuntu Vagrant Box
   ```bash
   $ vagrant box add ubuntu/focal64
   ```
1. Replace `<ROOTDIR>` with this directory in the `config.vm.synced_folder` property in `deployments/ubuntu1/Vagrantfile` and `deployments/ubuntu2/Vagrantfile`. Example: if this directory is `/home/user/go/src/github.com/spire-tutorials/docker-compose/delegatedidentity`, the property should look like:
```
  config.vm.synced_folder = "/home/user/go/src/github.com/spire-tutorials/docker-compose/delegatedidentity", "/delegatedidentity"
```

## Create Virtual Machines
To build virtual machines that run the services, run:
```bash
$ make env
```
Note that it can take 11+ minutes to bring up the VMs, install packages, and build and deploy services.
## Inspect Container Logs
### Callee Logs
```bash
$ ./scripts/callee-logs.sh
```

### Caller Logs
```bash
$ ./scripts/caller-logs.sh
```

### Proxy Logs
```bash
$ ./scripts/proxy-logs.sh
```

## Interact with Containers
To run SPIRE Server commands, use the `scripts/spire-server.sh` script, e.g.:
```bash
$ ./scripts/spire-server.sh entry show
Found 4 entries
Entry ID         : 809ffd14-cf91-49bd-9507-6f943f39d4b6
SPIFFE ID        : spiffe://example.org/callee
Parent ID        : spiffe://example.org/spire/agent/cn/ubuntu1
Revision         : 0
TTL              : default
Selector         : docker:label:org.example.service:callee

Entry ID         : 68ae0d12-d605-49f6-ab51-d520af60ca84
SPIFFE ID        : spiffe://example.org/caller1
Parent ID        : spiffe://example.org/spire/agent/cn/ubuntu2
Revision         : 0
TTL              : default
Selector         : docker:label:org.example.service:caller1

Entry ID         : f95431fa-8f3e-43ad-8f91-04eae1485366
SPIFFE ID        : spiffe://example.org/caller2
Parent ID        : spiffe://example.org/spire/agent/cn/ubuntu2
Revision         : 0
TTL              : default
Selector         : docker:label:org.example.service:caller2

Entry ID         : 80345870-76c1-4ae5-bd04-d04258cc66a7
SPIFFE ID        : spiffe://example.org/proxy
Parent ID        : spiffe://example.org/spire/agent/cn/ubuntu2
Revision         : 0
TTL              : default
Selector         : docker:label:org.example.service:proxy
```

To log into the VM hosting SPIRE Server, SPIRE Agent, and the `callee` service:
```bash
$ cd deployments/ubuntu1 && vagrant ssh
```

To log into the VM hosting SPIRE Agent, `caller1`, `caller2`, `caller3`, and `proxy`:
```bash
$ cd deployments/ubuntu2 && vagrant ssh
```

## Tear down environment
To tear down the virtual machines created with `make env`, run
```bash
$ make cleanenv
```

## Troubleshooting
### Vagrant failing at SSH to VirtualBox VM
Try restarting your machine and trying again.
