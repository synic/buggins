#!/bin/bash

set -e

if [ "$RUN_MIGRATIONS" = "1" ]; then
    yarn typeorm:cli -c default migration:run
fi

yarn start:prod
