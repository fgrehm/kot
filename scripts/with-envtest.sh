#!/usr/bin/env bash

set -o errexit
set -o pipefail

maj_minor=${K8S_VERSION%.*}
maj_minor=${maj_minor:1}
source <(${ENVTEST} use --arch amd64 -p env "${maj_minor}")

$*
