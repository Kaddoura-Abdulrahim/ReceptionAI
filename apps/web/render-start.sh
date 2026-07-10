#!/bin/sh
set -eu

HOSTNAME=0.0.0.0 exec node apps/web/server.js
