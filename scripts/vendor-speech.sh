#!/usr/bin/env bash
# Copies the standalone build of @plexusone/speech into the review page's
# static files, where it is embedded in the vac binary.
#
#   scripts/vendor-speech.sh 0.1.0            # from the npm registry
#   SPEECH_DIST=path/to/dist scripts/vendor-speech.sh 0.1.0   # from a local build
set -euo pipefail

version="${1:?usage: vendor-speech.sh VERSION}"
out="pkg/review/static/vendor/speech.js"
file="speech.global.js"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

if [[ -n "${SPEECH_DIST:-}" ]]; then
  src="$SPEECH_DIST/$file"
else
  npm pack "@plexusone/speech@$version" --pack-destination "$work" --silent >/dev/null
  tar -xzf "$work"/plexusone-speech-"$version".tgz -C "$work"
  src="$work/package/dist/$file"
fi

if [[ ! -f "$src" ]]; then
  echo "vendor-speech: $src not found" >&2
  exit 1
fi
if grep -qi '</script' "$src"; then
  echo "vendor-speech: $src contains </script and cannot be inlined" >&2
  exit 1
fi

mkdir -p "$(dirname "$out")"
{
  printf '/*! @plexusone/speech %s (MIT). Copied by "make vendor-speech"; do not edit. */\n' "$version"
  cat "$src"
} >"$out"
echo "vendor-speech: wrote $out (@plexusone/speech $version)"
