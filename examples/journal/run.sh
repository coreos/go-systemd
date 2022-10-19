#!/bin/bash

set -e

go build

echo "Running directly"
./journal

echo "Running through systemd"
unit_name="run-$(systemd-id128 new)"
systemd-run -u "$unit_name" --user --wait --quiet ./journal
journalctl --user -u "$unit_name"
