# Kubernetes CRD & controller for building a website in one step

## Goal

Inspired by [Kubernetes in Action ](), I can easily build up a website via a git repo and nginx /w kubernetes workload

## Development
1. define ur crd struct
2. generate the controller stuff
3. write ur event handler

## code generation

```bash
➜  my-crd-controller ./codegen.sh 
Generating deepcopy funcs
Generating clientset for mycontroller:v1alpha1 at github.com/nevermosby/my-crd-controller/pkg/client/clientset
Generating listers for mycontroller:v1alpha1 at github.com/nevermosby/my-crd-controller/pkg/client/listers
Generating informers for mycontroller:v1alpha1 at github.com/nevermosby/my-crd-controller/pkg/client/informers

```