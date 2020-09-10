{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "m4d-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "m4d-proxy.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "m4d-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "m4d-proxy.labels" -}}
helm.sh/chart: {{ include "m4d-proxy.chart" . }}
{{ include "m4d-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "m4d-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "m4d-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "m4d-proxy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "m4d-proxy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the dns search name to use
*/}}
{{- define "m4d-proxy.search" -}}
{{- printf "%s.%s" .Release.Namespace "svc.cluster.local" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the instance to use
*/}}
{{- define "m4d-proxy.instance" -}}
{{- printf .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the fqdn to use
*/}}
{{- define "m4d-proxy.fqdn" -}}
{{- printf "%s.%s" (include "m4d-proxy.instance" .) (include "m4d-proxy.search" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the egressgateway to use
*/}}
{{- define "m4d-proxy.egressgateway" -}}
{{- if .Values.proxy.egressGateway.usePrivate }}
{{- printf "%s-%s" (include "m4d-proxy.instance" .) "egressgateway" | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "istio-egressgateway" }}
{{- end }}
{{- end }}

{{/*
Create the name of the long egressgateway to use
*/}}
{{- define "m4d-proxy.egressgatewayFqdn" -}}
{{- if .Values.proxy.egressGateway.usePrivate }}
{{- printf "%s.%s" (include "m4d-proxy.egressgateway" .) (include "m4d-proxy.search" .) | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s.%s" (include "m4d-proxy.egressgateway" .) "istio-system.svc.cluster.local" | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
