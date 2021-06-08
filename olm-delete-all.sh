#!/bin/bash

kubectl delete subscription podset-operator -n podset-system
kubectl delete clusterserviceversion --all -n podset-system
kubectl delete crd podsets.app.example.com
kubectl delete operator podset-operator.podset-system
kubectl delete operatorgroup podset-operator -n podset-system
kubectl delete catalogsource podset-operator -n podset-system
