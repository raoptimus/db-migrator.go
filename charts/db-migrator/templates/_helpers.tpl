{{/*
Expand the name of the chart.
*/}}
{{- define "db-migrator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncated at 63 chars because some Kubernetes name fields are limited to this.
*/}}
{{- define "db-migrator.fullname" -}}
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
Chart name and version as used by the chart label.
*/}}
{{- define "db-migrator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "db-migrator.labels" -}}
helm.sh/chart: {{ include "db-migrator.chart" . }}
{{ include "db-migrator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "db-migrator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "db-migrator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Name of the service account to use.
*/}}
{{- define "db-migrator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "db-migrator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Fully qualified image reference. An explicit .Values.image.tag wins; otherwise the
tag falls back to "<appVersion>-alpine", matching the tags published to Docker Hub
(the registry only carries "-alpine" tags, no bare version tags).
*/}}
{{- define "db-migrator.image" -}}
{{- $tag := .Values.image.tag | default (printf "%s-alpine" .Chart.AppVersion) -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end }}

{{/*
Container env block shared by the release and rollback Jobs.
DSN is required and pulled from an existing Secret via secretKeyRef.
INTERACTIVE is hard-coded to "false": a Job has no TTY.
*/}}
{{- define "db-migrator.env" -}}
{{- $secret := required "migrator.dsn.existingSecret is required: provide the name of a Secret holding the DSN" .Values.migrator.dsn.existingSecret -}}
- name: DSN
  valueFrom:
    secretKeyRef:
      name: {{ $secret }}
      key: {{ .Values.migrator.dsn.secretKey }}
- name: MIGRATION_PATH
  value: {{ .Values.migrator.path | quote }}
- name: MIGRATION_TABLE
  value: {{ .Values.migrator.table | quote }}
{{- if .Values.migrator.clusterName }}
- name: MIGRATION_CLUSTER_NAME
  value: {{ .Values.migrator.clusterName | quote }}
{{- end }}
- name: MIGRATION_REPLICATED
  value: {{ .Values.migrator.replicated | quote }}
- name: MAX_CONN_ATTEMPTS
  value: {{ .Values.migrator.maxConnAttempts | quote }}
- name: COMPACT
  value: {{ .Values.migrator.compact | quote }}
- name: INTERACTIVE
  value: "false"
- name: DRY_RUN
  value: {{ .Values.migrator.dryRun | quote }}
{{- if .Values.migrator.placeholderCustom }}
- name: PLACEHOLDER_CUSTOM
  value: {{ .Values.migrator.placeholderCustom | quote }}
{{- end }}
{{- with .Values.migrator.extraEnv }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{/*
Shared pod spec for the migration Jobs.
Usage: include "db-migrator.podSpec" (dict "root" . "command" "release")
*/}}
{{- define "db-migrator.podSpec" -}}
{{- $root := .root -}}
{{- $command := .command -}}
restartPolicy: {{ $root.Values.restartPolicy }}
{{- with $root.Values.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{- end }}
serviceAccountName: {{ include "db-migrator.serviceAccountName" $root }}
{{- with $root.Values.podSecurityContext }}
securityContext:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with $root.Values.initContainers }}
initContainers:
  {{- toYaml . | nindent 2 }}
{{- end }}
containers:
  - name: db-migrator
    image: {{ include "db-migrator.image" $root }}
    imagePullPolicy: {{ $root.Values.image.pullPolicy }}
    args:
      - {{ $command | quote }}
    {{- with $root.Values.securityContext }}
    securityContext:
      {{- toYaml . | nindent 6 }}
    {{- end }}
    env:
      {{- include "db-migrator.env" $root | nindent 6 }}
    {{- with $root.Values.migrator.extraEnvFrom }}
    envFrom:
      {{- toYaml . | nindent 6 }}
    {{- end }}
    {{- with $root.Values.extraVolumeMounts }}
    volumeMounts:
      {{- toYaml . | nindent 6 }}
    {{- end }}
    {{- with $root.Values.resources }}
    resources:
      {{- toYaml . | nindent 6 }}
    {{- end }}
{{- with $root.Values.extraVolumes }}
volumes:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with $root.Values.nodeSelector }}
nodeSelector:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with $root.Values.tolerations }}
tolerations:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with $root.Values.affinity }}
affinity:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}
