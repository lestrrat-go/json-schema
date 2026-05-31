#!/bin/bash

# Run Go native fuzz targets for a single package.
#
# Usage:
#   ./fuzz.sh <package>      # e.g. ./fuzz.sh .   or   ./fuzz.sh ./validator
#
# Every Fuzz* target discovered in the package is exercised in turn. Duration per
# target is controlled by the FUZZTIME environment variable (default 5m); CI passes
# a short value on pull requests for a smoke run and the full value on schedule.
#
# This script cd's to its own directory so it can be invoked from anywhere.

set -euo pipefail

cd "$(dirname "$0")"

pkg="${1:-}"
if [[ -z "$pkg" ]]; then
	echo "usage: $0 <package>   (e.g. . or ./validator)" >&2
	exit 2
fi

fuzztime="${FUZZTIME:-5m}"

# Discover fuzz targets in the package. `-list` prints matching identifiers plus a
# trailing package-result line; grep keeps only the Fuzz* names.
targets="$(go test "$pkg" -list '^Fuzz' -run '^$' | grep '^Fuzz' || true)"

if [[ -z "$targets" ]]; then
	echo "no fuzz targets found in $pkg"
	exit 0
fi

status=0
for target in $targets; do
	echo "==> fuzzing ${target} in ${pkg} (fuzztime=${fuzztime})"
	if ! go test "$pkg" -run '^$' -fuzz "^${target}\$" -fuzztime "$fuzztime"; then
		status=1
	fi
done

exit "$status"
