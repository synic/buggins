#!/bin/sh

yarn typeorm:cli migration:generate ./src/databases/migrations/default/$@
