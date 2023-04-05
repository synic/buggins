#!/bin/sh

RESTART_TIMEOUT="${SERVICE_RESTART_TIMEOUT:-2}"
DATABASE_CONTAINER="${DATABASE_CONTAINER:-buggins-db}"

if [ "$NODE_ENV" = "development" ]; then
    echo -n "Waiting for db to become available... "
    ./docker/wait-for.sh -t 0 $DATABASE_CONTAINER:5432
    sleep 5
    echo "ready."
fi

while true; do
    eval $@
    RETCODE=$?

    if [ "$RESTART_TIMEOUT" = "skip" ]; then
        exit $RETCODE
    fi

    echo "Process died. Sleeping ${RESTART_TIMEOUT}s before restarting."
    sleep $RESTART_TIMEOUT
done
