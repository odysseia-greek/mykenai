---
# Source: {{.Name}}/templates/job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.Name}}
  namespace: odysseia
  labels:
    heritage: "Helm"
    release: "{{.Name}}"
    chart: "{{.Name}}"
    env: localdev
    variant: k3d
    app: {{.Name}}
spec:
  template:
    metadata:
      labels:
        app: {{.Name}}
        release: {{.Name}}
        version: v0.0.11
      annotations:
        odysseia-greek/role: seeder
        odysseia-greek/access: tracing
        perikles/accesses: solon
    spec:
      initContainers:
        - name: "periandros"
          image: ghcr.io/odysseia-greek/periandros:v0.0.11

          imagePullPolicy: Always
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: ELASTIC_ROLE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['odysseia-greek/role']
            - name: ELASTIC_ACCESS
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['odysseia-greek/access']
          envFrom:
            - configMapRef:
                name: euripides
          volumeMounts:
            - name: solon-certs
              mountPath: /etc/certs/solon
              readOnly: true
      containers:
        - name: "ptolemaios"
          image: ghcr.io/odysseia-greek/ptolemaios:v0.0.11

          env:
            - name: VAULT_SERVICE
              value: https://vault:8200
            - name: VAULT_TLS
              value:  "true"
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          envFrom:
            - configMapRef:
                name: euripides
          ports:
            - containerPort: 50051
          volumeMounts:
            - name: vault-server-tls
              mountPath: /etc/certs/vault
              readOnly: true
            - name: solon-certs
              mountPath: /etc/certs/solon
              readOnly: true
          imagePullPolicy: Always
          resources:
            requests:
              memory: 32Mi
              cpu: 50m
            limits:
              memory: 64Mi
              cpu: 100m
        - name: "{{.Name}}"
          image: ghcr.io/odysseia-greek/{{.Name}}:v0.0.11
          imagePullPolicy: Never

          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: ELASTIC_ACCESS
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['odysseia-greek/access']
          envFrom:
            - configMapRef:
                name: euripides
          ports:
            - containerPort: 2345
              name: delve
      restartPolicy: Never
      volumes:
        - name: vault-server-tls
          secret:
            secretName: vault-server-tls
        - name: solon-certs
          secret:
            secretName: solon-tls-certs
  backoffLimit: 3
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
spec:
  ports:
    - port: 2345
      name: delve
      targetPort: delve
  selector:
    app: {{.Name}}
