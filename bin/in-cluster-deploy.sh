#!/usr/bin/env bash
kubectl create configmap injector-config  --from-env-file=default.properties
kubectl apply -f bin/deployment.yaml