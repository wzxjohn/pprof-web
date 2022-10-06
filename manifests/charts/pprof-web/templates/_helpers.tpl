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
{{ include "pprof.web.selectorLabels" . }}
{{ if .Chart.AppVersion -}}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "pprof.web.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pprof.web.fullName" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

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