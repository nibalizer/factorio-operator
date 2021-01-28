# Factorio-Operator

This repo is an operator for kubernetes that manages servers for (factorio)[https://www.factorio.com/]

Generated by kubebuilder
Uses images from factoriotools


## Quick Start

```
# install kubebuilder
# verify k8s credentials
make install
make run
```

```
# Create a server
kubectl apply -f hack/extra.yaml
```


```
# view servers
$ kubectl get factorioservers
NAME                       PORT
factorioserver-00          32082
factorioserver-01          32066
factorioserver-02          30002
```

Connect to the server at `<node public ip>:<PORT>`

