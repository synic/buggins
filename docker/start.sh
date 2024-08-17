#!/bin/sh

cd /
./migrate -source file://./migrations -database sqlite://./data/database.sqlite up
./bot
