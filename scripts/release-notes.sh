#!/usr/bin/env bash
# Extract one release's notes from CHANGELOG.md for goreleaser --release-notes.
#
# Prefers the section matching the tag (## [0.2.0] - ...). Falls back to
# ## [Unreleased] so a release can be dry-run before the section is renamed;
# renaming Unreleased to the version is the release-prep step.
#
# Usage: scripts/release-notes.sh <tag> [changelog-path]
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <tag> [changelog-path]" >&2
  exit 2
fi

TAG="$1"
CHANGELOG="${2:-CHANGELOG.md}"
VERSION="${TAG#v}"

if [[ ! -f $CHANGELOG ]]; then
  echo "$0: $CHANGELOG not found" >&2
  exit 1
fi

# Print the body of the section whose heading is "## [<heading>]", stopping at
# the next "## " heading. Blank lines are held back and only flushed when more
# content follows, which trims the trailing blank lines before the next section.
extract_section() {
  local heading="$1"
  awk -v heading="$heading" '
    !inside {
      if ($0 == "## [" heading "]" || index($0, "## [" heading "] ") == 1) inside = 1
      next
    }
    /^## / { exit }
    /^[[:space:]]*$/ { if (started) pending++; next }
    {
      for (; pending > 0; pending--) print ""
      started = 1
      print
    }
  ' "$CHANGELOG"
}

NOTES="$(extract_section "$VERSION")"

if [[ -z ${NOTES//[[:space:]]/} ]]; then
  NOTES="$(extract_section "Unreleased")"
  if [[ -n ${NOTES//[[:space:]]/} ]]; then
    echo "$0: no '## [$VERSION]' section, using '## [Unreleased]'" >&2
  fi
fi

if [[ -z ${NOTES//[[:space:]]/} ]]; then
  echo "$0: no notes for $TAG in $CHANGELOG (looked for '## [$VERSION]' and '## [Unreleased]')" >&2
  exit 1
fi

printf '%s\n' "$NOTES"
