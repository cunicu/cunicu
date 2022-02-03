# Distributed Ledger Routing Protocol (DLRP)

This page descripes a protocol for distributing IP routing information and resource allocations in a distributed and untrusted environment.
In its core, the protocol is based on the ideas of a [_distributed ledger_][dlt] (a.k.a _blockchain_) and the [_web of trust_][wot] (WoT).

**Note:** This is a design draft.

## Web of Trust

Each agent defines a set of peers which it trusts.
Hereinafter, these trusted peers are referred to as _friends_.

The trust relationship of all agents is used to construct a _Web of Trust_ (WoT).
The WoT is a directed graph in which agents are represented by vertices and their trust relationship by an edge.

Each agent uses the WoT to extend the set of directly trusted _friends_ by the _friends_ of its _friends_.
This step can repeated multiple times up to a defined [degree of separation][6deg-of-separation]
We call this extended set of _friends_ the _neighburhood_ of an agent.

This degree of separation can be choosen by each agent individually and is a parameter for the willingness of an agent to connect with more distant peers.

Friendship is a directed property and as a result each agent has its own _neighborhood_. 
Agents only establish and accept connections to peers in its _neighborhood_.
Hence communication between to peers is only possible if they are both within the _neighborhood_ of each other.

## Trust ledger

Usually, _friends_ of an agent are variable.
An agent can befriend more peers or terminate its friendship with peers at any time.

In order for an agent to correctly determine its _neighborhood_ it requires an up to date view of the current WoT.
As the WoT is constructed from _friends_ of all peers, it is crucial for all nodes to have up-to-date knownledge about the _friends_ of each agent.
(To be exact, each agent only needs to know the _friends_ of its _friends_ and so on as each agent is only really interested in its own _neighborhood_.
So a full view of the WoT is not really required.)

The trust log is a distributed ledger which stores the set of _friends_ for each agent.
Changes in the set of _befriended_ peers are reflected by transactions on the ledger.

The ledger replicates these trust attestations to all other peers in the network.
As a result each agent which participates in the ledger can generate an up-to-date view of the WoT.

The following sections the basic properties of the trust ledger.

### Consensus

For the ledger to agree on the next transaction, a consensus algorithm is required.
Here, Proof-of-Elapsed-Time (PoET) is used as the consus algorithm of choice.

### Messaging

The consus algorithm requires a messaging transport for communicating new transactions with other peers.
Here the messaging layer is using the UDP transport protocol on port 27192.

Every agent listens on incoming datagrams on the port for Protobuf encoded messages of the following format.

## Link-local IPv6 addressing

WICE assigns each node a link-local IPv6 address which is calucated by the following formular:

    IPv6 link-local address = fe80:0:0:0 || SipHash64(X, k)

    where

        X is the public key of the corresponding Wireguard interface
        k is a the byte sequence 0x67, 0x67, 0x2c, 0x05, 0xd1, 0x3e, 0x11, 0x94, 0xbb, 0x38, 0x91, 0xff, 0x4f, 0x80, 0xb3, 0x97

[dlt]: https://en.wikipedia.org/wiki/Distributed_ledger

[wot]: https://en.wikipedia.org/wiki/Web_of_trust

[6deg-of-separation]: https://en.wikipedia.org/wiki/Six_degrees_of_separation
