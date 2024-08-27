profiles:
  - name: {{.Name}}
    build:
      artifacts:
        - image: ghcr.io/odysseia-greek/{{.Name}}
          context: ./{{.Name}}
          docker:
            target: debug
          sync:
            manual:
              - src: '**/*.go'
                dest: '/app'
    deploy:
      helm:
        releases:
          - name: alexandros
            chartPath: ../../odysseia-greek/mykenai/themistokles/odysseia/charts/{{.Name}}
            valuesFiles:
              - ../../odysseia-greek/mykenai/themistokles/odysseia/values/local.yaml
              - ../../odysseia-greek/mykenai/themistokles/odysseia/values/skaffold-values.yaml
            setValues:
              image.odysseiaapi.repo: {{.Name}}
              image.odysseiaapi.tag: dev
