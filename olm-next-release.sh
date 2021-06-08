#!/bin/bash

# FYI
# The `yq4` below is from https://github.com/mikefarah/yq (v4)

# IMPORTANT
# Change variable below
export NEXT_VERSION=0.0.4

export DOCKER_USERNAME=zulhfreelancer
export OPERATOR_IMG_NAME=podset-operator
export BUNDLE_IMG_NAME=podset-olm-bundle
export INDEX_IMG_NAME=podset-olm-index
export CSV_FILE_PATH=bundle/manifests/$OPERATOR_IMG_NAME.clusterserviceversion.yaml

CURRENT_VERSION=$(yq4 eval '.spec.version' $CSV_FILE_PATH)
echo "Current version : $CURRENT_VERSION"
echo "Next version    : $NEXT_VERSION"
if [ "$CURRENT_VERSION" == "$NEXT_VERSION" ]; then echo "Error: current and next version are equal"; exit 1; fi

OPERATOR_IMG=docker.io/$DOCKER_USERNAME/$OPERATOR_IMG_NAME:v$NEXT_VERSION
make docker-build docker-push IMG=$OPERATOR_IMG

make bundle IMG=$OPERATOR_IMG VERSION=$NEXT_VERSION
yq4 eval -i '.spec.replaces = "'$OPERATOR_IMG_NAME.v$CURRENT_VERSION'"' $CSV_FILE_PATH
BUNDLE_IMG=docker.io/$DOCKER_USERNAME/$BUNDLE_IMG_NAME:v$NEXT_VERSION
make bundle-build BUNDLE_IMG=$BUNDLE_IMG
make docker-push IMG=$BUNDLE_IMG

INDEX_IMG=docker.io/$DOCKER_USERNAME/$INDEX_IMG_NAME:latest
opm index add --bundles $BUNDLE_IMG --tag $INDEX_IMG --build-tool docker --from-index $INDEX_IMG
docker push $INDEX_IMG
