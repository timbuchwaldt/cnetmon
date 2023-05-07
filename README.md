- constant TCP stream
- new TCP connection
  - via DNS
  - via IP (apiserver)
- new UDP connection
  - via DNS
  - via IP (apiserver)
- ingress HTTP
- LB service
  - TCP
    - Via IP
    - Via DNS
  - UDP
    - Via IP
    - Via DNS


# Working principile

deployed as DS, part of service for discovery

## constant connections
state machine. new -> connected <-> disconnected

## spot-checks
worker pool