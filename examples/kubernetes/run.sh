#!/bin/bash

COMPONENT=${1:-state.memory}
COMPONENT_IMAGE=$(echo $COMPONENT |sed -e s/"\."/"-"/gi)
docker build -t $COMPONENT_IMAGE -f ../Dockerfile ../.. --build-arg COMPONENT=$COMPONENT --no-cache
docker tag $COMPONENT_IMAGE localhost:5001/$COMPONENT_IMAGE:latest
docker push localhost:5001/$COMPONENT_IMAGE:latest
kubectl apply -f ../$COMPONENT/component.yml
cat deployment.yaml | sed s/__COMPONENT__/$COMPONENT_IMAGE/gi | kubectl apply -f -