#!/bin/bash

# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

thishost=$(ip route get 1.1.1.1 | grep -oP 'src \K\S+')
echo ${thishost}

mkdir -p ./.vscode

FILE=./.vscode/settings.json
if test -f "$FILE"; then
    echo "$FILE already exists, cowardly exiting script."
    exit
fi

cat  <<EOF > $FILE
{
    "rest-client.environmentVariables": {

        "$shared": {
            "demo": "/debug"
        },
        "dev":{
            "host" :"http://${thishost}:8181"
        },
        "staging":{
            "host" :"http://${thishost}:8181"
        },
        "prod":{
            "host" :"http://${thishost}:8181"
        }
    }
}
EOF
