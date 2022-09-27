#!/bin/bash
COMPONENT=${1:-state.memory}

cp $COMPONENT/component.yml components/.

COMPONENT=$COMPONENT docker-compose build $ARGS
COMPONENT=$COMPONENT docker-compose up