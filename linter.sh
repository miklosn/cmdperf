#!/bin/bash

# Run it twice.
#
# First attempt may reformat/modify files, and therefore exit with non-zero status.
#
# Second attempt will not do anything and exit 0 unless there's a real problem beyond
# the code formatting that was completed.

pre-commit run --files $* >/dev/null \
    || pre-commit run --files $*

