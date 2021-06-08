## PodSet Kubernetes Operator

A basic k8s operator for PodSet CRD, built with Go and Operator SDK

## OLM Integration

Check-out [my OLM notes](olm-notes.md) if you want to learn how to publish/release your operator using [Operator Lifecycle Manager (OLM)](https://olm.operatorframework.io) project 

## Resources

- [https://learn.openshift.com/operatorframework/go-operator-podset/](https://learn.openshift.com/operatorframework/go-operator-podset/)
- [https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/](https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/)

## Commands

- To initialize a new project:

  `operator-sdk init --domain=example.com --repo=github.com/redhat/podset-operator`

  Note: `github.com/redhat/podset-operator` is the Go project module name (see `go.mod` file)

- To create API and controller:

  `operator-sdk create api --group=app --version=v1alpha1 --kind=PodSet --resource --controller`

- After modifying the `*_types.go` file, run:

  `make generate`

- To generate CRD:

  `make manifests`

- To run the operator controller locally (assumming `kubeconfig` was configured):

  `WATCH_NAMESPACE=myproject make run`
