#!/bin/sh

set -eux

if [ "$1" = 'xtracego' ]; then
    exec "$@"
else
    exec xtracego "$@"
fi