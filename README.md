# INTRODUCTION

This repo contains the source code of Kaloom's Kubernetes `null` ipam cni-plugin

it's meant to be used in use cases where the ipv4/ipv6 address configuration of a network device in a Pod is the responsibility of the application rather than the `ipam` cni-plugin (e.g. vrouters) for auxiliary network devices when multiple network attachment is used, see [kactus cni-plugin](https://github.com/kaloom/kubernetes-kactus-cni-plugin)

## Example configuration

```
{
  "name": "null-net",
  "type": "null"
}
```

## Network configuration reference

* `name` (string, required): the name of the network.
* `type` (string, required): "null".

# HOW TO BUILD

> `./build.sh`

## For developpers:

This project uses Go modules as dependency management system. Even if modules support was added in 1.11, the minimum Go version required by this project is 1.13.

If you're adding a new dependency package to the project you need to update go.mod and go.sum files. Refer to this Go [blog post](https://blog.golang.org/using-go-modules) or this [wiki article](https://github.com/golang/go/wiki/Modules) if you need information on how to work with go modules.

# Setup

How to deploy `null` cni-plugin

## install in `/opt/cni/bin`

1. depoly the `null` cni-plugin by simply copying the built artifact in `bin/null` to the cni bin directory (i.e. typically under `/opt/cni/bin`) and that for every node in kubernetes cluster

> $ `sudo cp bin/null /opt/cni/bin`

## As DaemonSet

1. delopy the null cni-plugin as a daemon set:

> $ `kubectl apply -f manifests/null-ds.yaml`

### Note
Currently, to deploy kactus as DaemonSet
* *selinux* should not be in *enforced* mode (*permissive* mode is okay):
  > \# `setenforce permissive`

  > \# `sed -i 's/^SELINUX=.*/SELINUX=permissive/g' /etc/selinux/config`

# Example

see [kactus cni-plugin Example](https://github.com/kaloom/kubernetes-kactus-cni-plugin#example)