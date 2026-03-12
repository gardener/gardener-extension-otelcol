{{/*
A common name for fluent-bit resources
*/}}
{{- define "fluent-bit.name" -}}
{{- default "fluent-bit" .Values.fluentbit.name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels for fluent-bit resources
*/}}
{{- define "fluent-bit.labels" -}}
app.kubernetes.io/name: {{ include "fluent-bit.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Values.fluentbit.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels for fluent-bit resources
*/}}
{{- define "fluent-bit.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fluent-bit.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Fluent Bit Operator enabled label for selector matching
*/}}
{{- define "fluent-bit.enabledLabel" -}}
{{ include "fluent-bit.name" . }}.fluent.io/enabled: "true"
{{- end }}

{{/*
Common labels for fluent-bit resources
*/}}
{{- define "fluent-bit.annotations" -}}
{{- with .Values.fluentbit.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}