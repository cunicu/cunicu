# Session signalling

Lets assume two Wireguard peers `Pa` & `Pb` are seeking to establish a ICE session.

The smaller public key (PK) of the two peers takes the role of the controlling agent.
In this example PA is the controlling agent: PK(PA) < PK(PB). 

```
PA                         PB

  --- initial offer     -->        id=SID_Pa, version=0, candidates=[], eoc=false
  <-- initial offer     ---        id=SID_Pb, version=0, candidates=[], eoc=false

  --- subsequent offers -->        id=SID_Pa, version=1, candidates=[C1_Pa], eoc=false
  <-- subsequent offers ---        id=SID_Pb, version=1, candidates=[C1_Pb], eoc=false

  --- subsequent offers -->        id=SID_Pa, version=2, candidates=[C1_Pa, C2_Pa], eoc=false
  <-- subsequent offers ---        id=SID_Pb, version=2, candidates=[C1_Pb, C2_Pb], eoc=false

  ---  eoc. offer       -->        id=SID_Pa, version=3, candidates=[C1_Pa, C2_Pa], eoc=true
  <--  eoc. offer       ---        id=SID_Pb, version=3, candidates=[C1_Pb, C2_Pb], eoc=true
```

## Restart

Agent will restart
  - if
    - `last_recv.id` has been set
    - `recv.id!=last_recv.id`
    - `recv.version==0`

  - then
    - set
      - `local.id=rand()`
      - `local.version=0`
      - `local.candidates=[]`
    - publish new offer
    - wait for first offer including candidates from remote
    - (re)start agent
    - add first received 
    - start gathering candidates
      - send an offers for each candidate `c`:
        - `candidates=local.candidates.append(c)`
        - `id=local.id`
        - `rid=local.rid`
        - `version=local.version++`
 
## Offer

Offers are encoded as JSON:

```json
{
  "id": 1232353452,                                     // Unique session id
  "version": 0,                                         // Session version, incremented with each updated offer
  "cands": [                                            // List of ICE candidates
    {
      "type": "host",
      "foundation": "1742129347",
      "component": 1,
      "network": "udp4",
      "priority": 2130706431,
      "address": "10.2.0.11",
      "port": 37518
    }
  ],
  "eoc": false                                          // Flag to indicate that all candidates have been gathered (ICE trickle)
}
```
