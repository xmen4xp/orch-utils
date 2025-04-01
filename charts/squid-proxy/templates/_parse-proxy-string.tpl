# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

{{- define "squid-proxy.addressFromProxyString" }}
{{- $urlWithoutScheme := trimPrefix "http://" .Values.httpsProxy }}
{{- $parts := splitList ":" $urlWithoutScheme }}
{{- $domain := index $parts 0 }}
{{- printf "%s" $domain }}
{{- end }}

{{- define "squid-proxy.portFromProxyString" }}
{{- $urlWithoutScheme := trimPrefix "http://" .Values.httpsProxy }}
{{- $parts := splitList ":" $urlWithoutScheme }}
{{- $port := index $parts 1 }}
{{- printf "%s" $port }}
{{- end }}
