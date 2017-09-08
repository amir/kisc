# kisc
Kubernetes Initializers Sans Controllers

```
$ curl --data-binary "@./example.yaml" localhost:8080/evaluate
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  creationTimestamp: null
  initializers:
    pending: []
spec:
  replicas: 1
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: example
      name: example
    spec:
      containers:
      - image: example
        name: example
        resources: {}
      initContainers:
      - image: init
        name: init
        resources: {}
      volumes:
      - name: test
status: {}
```
