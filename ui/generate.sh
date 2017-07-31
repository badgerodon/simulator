#!/usr/bin/env bash

set -ex

gopherjs build -o "dist/bundle.js" .
cat "dist/bundle.js" | head -n -1 > "dist/tmp.js"
mv "dist/tmp.js" "dist/bundle.js"
rm "dist/bundle.map"
