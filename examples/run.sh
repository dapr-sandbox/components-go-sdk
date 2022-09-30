#!/bin/bash
COMPONENT=${1:-state.memory}

if [ ! -d "components" ]; then
    mkdir components
fi

ADDITIONAL_ARGS=
if [ -e $COMPONENT/docker-compose.dependencies.yml ]; then
    ADDITIONAL_ARGS="-f $COMPONENT/docker-compose.dependencies.yml"
fi
cp $COMPONENT/component.yml components/.

COMPONENT=$COMPONENT docker-compose -f docker-compose.yml $ADDITIONAL_ARGS build $ARGS
COMPONENT=$COMPONENT docker-compose -f docker-compose.yml $ADDITIONAL_ARGS up