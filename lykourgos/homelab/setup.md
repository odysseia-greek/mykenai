# HOMELAB

## setup

### first steps

Get all the raspis ready. Cluster them and start burning images.

Add static ips for each of the nodes

### bootstrapping

run ansible script to get all the pies ready to run k3s

setup ansible script for bootstrapping k3s according to the following params:

1. longhorn
2. cilium
3. traefik
4. etcd


### setup


[servers]
k3s-s-athenai rpi5 8gb
k3s-s-sparta rpi4 8gb
k3s-s-syrakousai rpi5 8gb

[workers]
k3s-w-thebai rpi4 8gb
k3s-w-korinth rpi5 4gb
k3s-w-argos rpi4 4gb
k3s-w-taras rpi4 4gb