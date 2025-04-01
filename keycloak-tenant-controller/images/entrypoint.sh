#!/bin/sh

# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -e

LogLevel="${KTC_SERVER_LOG_LEVEL:-info}"
ClientType="${KC_CLIENT_TYPE:-http}"

exec ktc -loglevel="$LogLevel" &
pid=$!
trap 'kill -TERM $pid; wait $pid' TERM
wait $pid
