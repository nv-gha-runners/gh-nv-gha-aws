#!/bin/bash

set -euo pipefail

TAG=$1

mise build --tag="$TAG"
