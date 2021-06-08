#!/bin/bash

# IMPORTANT
# Don't use this script to publish new update for operator.
# Use olm-next-release.sh file instead.
#
# Script below was modified based on my script here:
# https://pastebin.com/raw/dwgAkPtg

export DOCKER_USERNAME=zulhfreelancer
export OPERATOR_IMG_NAME=podset-operator
export BUNDLE_IMG_NAME=podset-olm-bundle
export INDEX_IMG_NAME=podset-olm-index

OPERATOR_IMG=docker.io/$DOCKER_USERNAME/$OPERATOR_IMG_NAME:v0.0.1
make docker-build docker-push IMG=$OPERATOR_IMG

make bundle IMG=$OPERATOR_IMG
BUNDLE_IMG=docker.io/$DOCKER_USERNAME/$BUNDLE_IMG_NAME:v0.0.1
make bundle-build BUNDLE_IMG=$BUNDLE_IMG
make docker-push IMG=$BUNDLE_IMG

INDEX_IMG=docker.io/$DOCKER_USERNAME/$INDEX_IMG_NAME:latest
opm index add --bundles $BUNDLE_IMG --tag $INDEX_IMG --build-tool docker
docker push $INDEX_IMG
