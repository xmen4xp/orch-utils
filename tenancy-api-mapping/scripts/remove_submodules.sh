#!/bin/bash

# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Determine the root directory of the repository
REPO_ROOT=$(git rev-parse --show-toplevel)

git reset -- "$REPO_ROOT/.gitmodules" "$REPO_ROOT/gitsubmodules/"

# Read each submodule path from .gitmodules and remove them
while IFS= read -r line; do
    if [[ $line =~ \[submodule\ \"(.+)\"\] ]]; then
        MODULE_PATH="${BASH_REMATCH[1]}"
        echo "Removing submodule at path: $MODULE_PATH"

        # Deinitialize the submodule
        git submodule deinit -f -- "$MODULE_PATH"

        # Remove the submodule from the index
        git rm --cached -f -- "$MODULE_PATH"

        # Remove the submodule's entry from .gitmodules
        git config -f "$REPO_ROOT/.gitmodules" --remove-section "submodule.$MODULE_PATH"

        # Manually remove the submodule's directory from .git/modules
        rm -rf "$REPO_ROOT/.git/modules/$MODULE_PATH"

        # Remove the actual submodule directory
        rm -rf "$MODULE_PATH"
    fi
done < "$REPO_ROOT/.gitmodules"

# Truncate the .gitmodules file to clear its contents
> "$REPO_ROOT/.gitmodules"

echo "All submodules have been removed."