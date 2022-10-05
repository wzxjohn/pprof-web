{{/* vim: set filetype=mustache: */}}

{{/*
Full name for pprof-web
*/}}
{{- define "pprof.web.fullName" -}}
{{- if .Values.fullNameOverride -}}
{{- .Values.fullNameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- if contains .Values.name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name .Values.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "pprof.web.labels" -}}
{{- if .Chart.AppVersion -}}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end -}}

{{/*
Service name for portal
*/}}
{{- define "pprof.web.serviceName" -}}
{{- if .Values.service.fullNameOverride -}}
{{- .Values.service.fullNameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{ include "pprof.web.fullName" .}}
{{- end -}}
{{- end -}}