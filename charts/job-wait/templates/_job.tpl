# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Wait for the completion of all jobs that are (1) in current namespace and (2) has a name starting with .Values.jobPrefix
{{- define "job-wait-contents" -}}
set -e;

jobs=""
until [ -n "$jobs" ]; do
  jobs=$(kubectl get jobs -n {{ .Release.Namespace }} -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | grep -E "^{{ required "A valid jobPrefix entry required!" .Values.jobPrefix }}")
  sleep 1
done
echo "Found jobs with prefix {{.Values.jobPrefix}}: $(echo $jobs |tr '\n' ' ')"

for job in $jobs; do
  echo "Waiting for job $job to complete"
  until kubectl wait --for=condition=complete --timeout=1s job/$job -n {{ .Release.Namespace }}; do
    sleep 1
  done
  echo "Job $job is completed"
done;
{{- end -}}
