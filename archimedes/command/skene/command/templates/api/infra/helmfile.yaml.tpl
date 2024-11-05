  - name: {{.Name}}
    namespace: odysseia
    chart: ./charts/{{.Name}}
    version: 0.1.1
    missingFileHandler: Error
    labels:
      tier: backend
    needs:
      - odysseia/solon
    values:
      - values/{{ .Environment.Name }}.yaml
      - images:
          odysseiaapi:
            repo: {{.Name}}
            tag: v0.0.14
    set:
      - name: replicas
        value: {{ .Values.{{.Name}}.replicas }}