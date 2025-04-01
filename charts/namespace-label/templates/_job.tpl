# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

{{- define "job-script-contents" -}}
set -e;
{{- range $ns := .Values.namespaces }}
kubectl create namespace {{ $ns.name }} --dry-run=client -o yaml | kubectl apply -f -;
{{- range $k, $v := $ns.labels }}
kubectl label namespace {{ $ns.name }} {{ $k }}={{ $v }} --overwrite;
{{- end }}
{{- end }}
echo "Namespaces have been processed successfully.";
{{- end -}}
