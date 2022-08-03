# TODOs

Please also have a look at the current [GitHub issues](https://github.com/stv0g/wice/issues) of the project.

-   Investigate distributed management of Access Control Lists (ACL)
    -   <https://ieeexplore.ieee.org/document/1437269>
    -   <https://www.it.iitb.ac.in/~madhumita/access/gcs/A%20Trust%20based%20Access%20Control%20Framework%20for%20P2P%20File%20Sharing%20Systems.pdf>
    -   <https://www.springerprofessional.de/en/decentralized-access-control-technique-with-multi-tier-authentic/19543988>
    -   <https://link.springer.com/chapter/10.1007%2F978-3-319-28865-9_28>

-   Single socket per WireGuard interface / ICE Agent

-   Update proxy instances instead of recreating them.
    -   Avoids possible packet loss during change of candidate pairs

-   Add better proxy implementations for OpenBSD, FreeBSD, Android and Windows

-   Add sub-commands for controlling `wice` daemon:
    -   `wice show [[INTF] [PEER]]`
    -   `wice add INTF`
    -   `wice delete INTF`
    -   `wice discover INTF GROUP`
    -   `wice sync [INTF]`
    -   `wice restart INTF PEER`
    -   `wice monitor`

-   Add check for handshakes before attempting to ping

-   Add context for waiting for events

-   Use mermaid actor diagram for signaling docs

-   Use RTT & packet loss for selecting ICE relay candidates

-   Move all the ToDo in this document to GitHub issues 