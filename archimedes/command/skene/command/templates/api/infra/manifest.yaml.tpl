---
# Source: {{.Name}}/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Name}}
  namespace: odysseia
data:
  ELASTIC_SEARCH_SERVICE: "https://aristoteles-es-http:9200"
  ENV: "localdev"
  TLS_ENABLED: "true"
  CERT_ROOT: "/etc/certs"
  SOLON_SERVICE: "https://solon:5443"
  ALEXANDROS_SERVICE: "http://alexandros:5000"
---
# Source: {{.Name}}/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
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
    domain: "backend"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
        release: {{.Name}}
        version: v0.0.11
      annotations:
        odysseia-greek/role: api
        odysseia-greek/access: text
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
                name: {{.Name}}
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
                name: {{.Name}}
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
        - name: "aristophanes"
          image: ghcr.io/odysseia-greek/aristophanes:v0.0.11

          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: ELASTIC_SEARCH_USER
              valueFrom:
                secretKeyRef:
                  name: agreus-elastic
                  key: user
            - name: ELASTIC_SEARCH_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: agreus-elastic
                  key: password
          envFrom:
            - configMapRef:
                name: {{.Name}}
          ports:
            - containerPort: 50052
          imagePullPolicy: Always
          resources:
            requests:
              memory: 32Mi
              cpu: 50m
            limits:
              memory: 64Mi
              cpu: 200m
          volumeMounts:
            - name: elastic-internal-tls
              mountPath: /app/config/elastic
              readOnly: true
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
                name: {{.Name}}
          volumeMounts:
            - name: solon-certs
              mountPath: /etc/certs/solon
              readOnly: true
          ports:
            - containerPort: 5000
              name: http
            - containerPort: 2345
              name: delve
      volumes:
        - name: vault-server-tls
          secret:
            secretName: vault-server-tls
        - name: elastic-internal-tls
          secret:
            secretName: aristoteles-es-http-certs-public
        - name: solon-certs
          secret:
            secretName: solon-tls-certs
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
spec:
  ports:
    - port: 5000
      name: http
      targetPort: http
    - port: 2345
      name: delve
      targetPort: delve
  selector:
    app: {{.Name}}
