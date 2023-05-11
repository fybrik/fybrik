{{/*
Expand the name of the chart.
*/}}
{{- define "fybrik.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fybrik.fullname" -}}
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
{{- define "fybrik.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "fybrik.labels" -}}
helm.sh/chart: {{ include "fybrik.chart" . }}
{{ include "fybrik.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "fybrik.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fybrik.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the value of an image field from hub, image and tag
*/}}
{{- define "fybrik.image" -}}
{{- $root := first . -}}
{{- $ctx := last . -}}
{{- if contains "/" $ctx.image }}
{{- printf "%s" $ctx.image }}
{{- else }}
{{- printf "%s/%s:%s" ( $ctx.hub | default $root.Values.global.hub ) $ctx.image ( $ctx.tag | default $root.Values.global.tag | default $root.Chart.AppVersion ) }}
{{- end }}
{{- end }}

{{/*
Extract the file name from a path
*/}}
{{- define "fybrik.opaServerPolicyFileName" -}}
{{- $path := toString (first .) -}}
{{- printf "%s" $path | base| toString }}
{{- end }}

{{/*
isEnabled evaluates an enabled flag that might be set to "auto".
Returns true if one of the following is true:
The return value when using `include` is always a String.
1. The flag is set to "true"
2. The flag is set to true
3. The flag is set to "auto" and the second parameter to this function is true 
*/}}
{{- define "fybrik.isEnabled" -}}
{{- $flag := toString (first .) -}}
{{- $condition := last . -}}
{{- if or (eq $flag "true") (and (eq $flag "auto") $condition) }}
true
{{- end -}}
{{- end }}

{{/*
isRazeeConfigurationEnabled checks if razee configuration is enabled
*/}}
{{- define "fybrik.isRazeeConfigurationEnabled" -}}
{{- if or .Values.coordinator.razee.user .Values.coordinator.razee.apiKey .Values.coordinator.razee.iamKey -}}
true
{{- end -}}
{{- end }}

{{/*
isArgocdConfigurationEnabled checks if argocd configuration is enabled
*/}}
{{- define "fybrik.isArgocdConfigurationEnabled" -}}
{{- if .Values.coordinator.argocd.user -}}
true
{{- end -}}
{{- end }}

{{/*
Detect the version of cert manager crd that is installed
Defaults to cert-manager.io/v1
*/}}
{{- define "fybrik.certManagerApiVersion" -}}
{{- if .Capabilities.APIVersions.Has "cert-manager.io/v1beta1" -}}
cert-manager.io/v1beta1
{{- else if .Capabilities.APIVersions.Has "cert-manager.io/v1alpha2" -}}
cert-manager.io/v1alpha2
{{- else if .Capabilities.APIVersions.Has "certmanager.k8s.io/v1alpha1" -}}
certmanager.k8s.io/v1alpha1
{{- else -}}
cert-manager.io/v1
{{- end -}}
{{- end -}}

{{/*
Get modules namespace
*/}}
{{- define "fybrik.getModulesNamespace" -}}
{{- .Values.modulesNamespace.name | default "fybrik-blueprints" -}}
{{- end -}}

{{/*
processPodSecurityContext does the following:
- skips certain keys in Values.global.podSecurityContext map if running on openshift
- merges Values.global.podSecurityContext with specific podSecurityContext settings
  that is passed as a parameter to this function, giving preference to the values in the
  latter map.
*/}}
{{- define "fybrik.processPodSecurityContext" }}
{{- $globalContext := deepCopy .context.Values.global.podSecurityContext  }}
{{- $podSecurityContext := .podSecurityContext }}
{{- if .context.Capabilities.APIVersions.Has "security.openshift.io/v1" }}
  {{- range $k, $v := .context.Values.global.podSecurityContext }}
    {{- if or (eq $k "runAsUser") (eq $k "seccompProfile") }}
      {{- $_ := unset $globalContext $k }}
    {{- end }}
   {{- end }}
{{- end }}
{{ mergeOverwrite $globalContext $podSecurityContext | toYaml }}
{{- end }}

{{/*
localChartsMountPath returns the mount path of a persistent volume
that can contain helm charts that can be referenced by the Fybrik module.
Relevant only when .Values.chartsPersistentVolumeClaim is set.
*/}}
{{- define "fybrik.localChartsMountPath" }}
{{- printf "/opt/fybrik" }}
{{- end }}

{{/*
Print Data directory.
*/}}
{{- define "fybrik.getDataDir" -}}
/data
{{- end }}

{{/*
Print sub directory in /data directory. The sub directory is
passed as parameter to the function.
*/}}
{{- define "fybrik.getDataSubdir" -}}
{{- $dir := toString (first .) -}}
{{- printf "%s/%s" (include "fybrik.getDataDir" .) $dir }}
{{- end }}
