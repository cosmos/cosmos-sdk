#!/bin/bash

# lint modified go files
golangci-lint run --fix --new -c .golangci.yml