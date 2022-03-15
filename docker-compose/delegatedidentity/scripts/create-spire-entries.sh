#!/bin/sh

function create_entry {
    /opt/spire/bin/spire-server entry create -socketPath /run/spire-server/private/api.sock "$@"
}

# Create entry for callee
create_entry -parentID 'spiffe://example.org/spire/agent/cn/ubuntu1' -spiffeID 'spiffe://example.org/callee' -selector 'docker:label:org.example.service:callee'

# Create entry for proxy so that it can access Delegated Identity APIs
create_entry -parentID 'spiffe://example.org/spire/agent/cn/ubuntu2' -spiffeID 'spiffe://example.org/proxy' -selector 'docker:label:org.example.service:proxy'

# Create entry for caller1
create_entry -parentID 'spiffe://example.org/spire/agent/cn/ubuntu2' -spiffeID 'spiffe://example.org/caller1' -selector 'docker:label:org.example.service:caller1'

# Create entry for caller2
create_entry -parentID 'spiffe://example.org/spire/agent/cn/ubuntu2' -spiffeID 'spiffe://example.org/caller2' -selector 'docker:label:org.example.service:caller2'
