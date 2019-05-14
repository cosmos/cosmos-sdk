#!/bin/bash

function term() {
    echo 'Caught SIGTERM, exiting.'
    exit 0
}

trap 'term' SIGTERM

while true
do
   sleep 1
done