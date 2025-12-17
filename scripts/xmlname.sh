#!/bin/sh
echo -n 'Testing if docs/aura-theme-dark.html has the wrong string color... '
if grep -q '.s { color: #00ffff; font-weight: bold' docs/aura-theme-dark.html; then
  echo FAIL
  exit 1
fi
echo OK
