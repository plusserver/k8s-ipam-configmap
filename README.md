# k8s-ipam-configmap

Stores IP addresses in a configmap.

## Deploy

Apply the custom resource:

```bash
kubectl apply -f deploy/crd.yaml
```

Edit the configmap in `deploy/configmap.yaml`, for example to configure the network you want to use. Then:

```bash
kubectl apply -f deploy/configmap.yaml
```

Then, apply the RBAC configuration and the controller itself:

```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deployment.yaml
```

The controller can be configured using the following configuration values:

- *LOG_LEVEL* (default: info) Log level (debug, info, ...)
- *IPAM_NETWORK* Use this network to serve addresses.
- *NAME_TEMPLATE* (default `{{.Tag}}.{{.Namespace}}.{{.Name}}`) How the address reservations are stored internally.

## How to use

To create an IP address reservation manually, create an IP address request:

```yaml
apiVersion: ipam.nexinto.com/v1
kind: IpAddress
metadata:
  name: myip
spec:
  description: My great service runs here.
```

Check using describe if the address was successfully assigned:

```bash
kubectl describe ipaddress myip
```

The Status fields should contain your address:

```yaml
...
Status:
  Address:   10.10.0.0
  Name:      ConfigMap.default.myip
  Provider:  ConfigMap
```

If something went wrong, an Event is created to explain why.

The IP addresses are namespaced.

The Spec supports the following fields:

- *description* (optional) description for this address reservation
- *name* (optional) name how the IP address management internally stores this address. The default is `ConfigMap.$NAMESPACE.$NAME`.
- *ref* (optional) do not create a new address; instead reuse an existing entry. Use the IPAM name (like `ConfigMap.$NAMESPACE.$NAME`), not the Kubernetes object name.

To list all addresses:

```bash
kubectl get ipaddress -o go-template='{{range .items}}{{.metadata.namespace}}-{{.metadata.name}} = {{.status.address}}{{"\n"}}{{end}}'
```