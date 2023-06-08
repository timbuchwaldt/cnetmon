# cnetmon - continuous network monitoring
cnetmon provides a way to continuously monitor your Kubernetes-based network. Deployed as a daemonset, the pods continuously resolve all other instances in the cluster (both via DNS SRV resolution as well as via the Kubernetes API) and do various connection tests like short-lived TCP connections (to check if newly established sessions work), long-lived TCP connections (to ensure they are optimally never interrupted) as well as UDP packets.

Metrics are taken from these tests and exported via the industry-standard Prometheus protocol.