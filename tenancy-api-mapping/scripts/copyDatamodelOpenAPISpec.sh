#!/bin/bash

# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Variables
REPO_HTTPS_URL="https://github.com/open-edge-platform/orch-utils.git"
LOCAL_REPO_DIR="temp_repo"
JSON_FILE_PATH="tenancy-datamodel/build/openapi/edge-orchestrator.intel.com.json"
OUTPUT_PATH="openapispecs/generated/orch-utils.tenancy-datamodel.openapi.yaml"
TAG="${DM_REPO_TAG_VERSION:-v1.0.19}"  # Use DM_REPO_TAG_VERSION from environment or default to v1.0.19

# Function to log messages
log() {
  echo "$(date +"%Y-%m-%d %H:%M:%S") - $1"
}

log "Starting the script."

# Temporarily disable detached HEAD advice
log "Disabling detached HEAD advice."
git config --global advice.detachedHead false

# Clone the repository using HTTPS and .netrc for authentication, and checkout the specific tag
log "Cloning the repository from $REPO_HTTPS_URL with tag $TAG."
git clone --branch $TAG --depth 1 $REPO_HTTPS_URL $LOCAL_REPO_DIR

# Check if the clone was successful
if [ $? -ne 0 ]; then
  log "Failed to clone the repository. Please check the HTTPS URL, the tag, and your .netrc file."
  # Re-enable detached HEAD advice
  git config --global advice.detachedHead true
  exit 1
fi
log "Repository cloned successfully."

# Check if the JSON file exists in the cloned repository
if [ ! -f "$LOCAL_REPO_DIR/$JSON_FILE_PATH" ]; then
  log "JSON file not found in the cloned repository. Please check the file path."
  # Clean up the cloned repository
  rm -rf $LOCAL_REPO_DIR
  # Re-enable detached HEAD advice
  git config --global advice.detachedHead true
  exit 1
fi
log "JSON file found at $LOCAL_REPO_DIR/$JSON_FILE_PATH."

# Convert JSON to YAML and save it to the specified location
log "Converting JSON to YAML."
cat "$LOCAL_REPO_DIR/$JSON_FILE_PATH" | yq -P > $OUTPUT_PATH

# Check if the conversion was successful
if [ $? -ne 0 ]; then
  log "Failed to convert JSON to YAML."
  # Clean up the cloned repository
  rm -rf $LOCAL_REPO_DIR
  # Re-enable detached HEAD advice
  git config --global advice.detachedHead true
  exit 1
fi
log "JSON file has been converted to YAML and saved to $OUTPUT_PATH."

# Remove PATCH operations from all paths
log "Removing PATCH operations from all paths."
yq eval 'del(.paths.[].patch)' -i $OUTPUT_PATH

# Check if the removal was successful
if [ $? -ne 0 ]; then
  log "Failed to remove PATCH operations from the OpenAPI YAML specification."
  # Clean up the cloned repository
  rm -rf $LOCAL_REPO_DIR
  # Re-enable detached HEAD advice
  git config --global advice.detachedHead true
  exit 1
fi
log "PATCH operations have been removed from the OpenAPI YAML specification."

# Add BearerAuth security to the OpenAPI YAML specification
log "Adding BearerAuth security to the OpenAPI YAML specification."
yq eval '.components.securitySchemes.BearerAuth = {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}' -i $OUTPUT_PATH
yq eval '.security = [{"BearerAuth": []}]' -i $OUTPUT_PATH

# Check if the addition was successful
if [ $? -ne 0 ]; then
  log "Failed to add BearerAuth security to the OpenAPI YAML specification."
  # Clean up the cloned repository
  rm -rf $LOCAL_REPO_DIR
  # Re-enable detached HEAD advice
  git config --global advice.detachedHead true
  exit 1
fi
log "BearerAuth security has been added to the OpenAPI YAML specification."

# Replace the servers field in the OpenAPI YAML specification
log "Replacing the servers field in the OpenAPI YAML specification."
yq eval '.servers = [{"url": "{apiRoot}", "variables": {"apiRoot": {"default": "https://<multitenancy-gateway-host>"}}}]' -i $OUTPUT_PATH

# Clean up the cloned repository
log "Cleaning up the cloned repository."
rm -rf $LOCAL_REPO_DIR

# Re-enable detached HEAD advice
log "Re-enabling detached HEAD advice."
git config --global advice.detachedHead true

log "Script completed successfully."
