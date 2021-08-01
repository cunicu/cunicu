# Backends

WICE can support multiple backends for signalling session information such as session IDs, ICE candidates, public keys and STUN credentials.

## Available backends

Currently HTTP REST, Kubernetes and libp2p are supported as backends.
Checkout the `Backend` interface in `wice/backend/backend.go` for implementing your own backend.
