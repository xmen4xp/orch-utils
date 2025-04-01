# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

{{- define "certificate.auto" -}}
{{- if and .Values.autoCert.enabled .Values.generateOrchCert -}}
{{- printf "Both autoCert and self-signed are set to true; defaulting to self-signed." -}}
false
{{- else if .Values.autoCert.enabled -}}
true
{{- else if .Values.generateOrchCert -}}
false
{{- else -}}
false
{{- end -}}
{{- end -}}