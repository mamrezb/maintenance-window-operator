# Maintenance Window Manager

A Kubernetes operator that monitors critical service endpoints to automatically indicate when a service is under maintenance. Frontend applications can query its lightweight HTTP API to decide whether to switch into maintenance mode.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Demo](#demo)
- [Quickstart](#quickstart)
- [Usage](#usage)
- [Development](#development)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

---

## Overview

Maintenance Window Manager continuously watches the health of services (via their Endpoints) in your Kubernetes cluster. By updating a custom resource’s status, it enables applications to react—such as displaying a maintenance page when a critical backend is unavailable.

---

## Architecture

Maintenance Window Manager is built using standard Kubernetes operator patterns:

- **Custom Resource (ServiceChecker):**  
  Declare the list of services to monitor, along with a flag indicating whether each is critical.

- **Controller:**  
  Watches for changes in both the ServiceChecker CRs and Kubernetes Endpoints. When a service’s state changes, the controller updates the CR’s status with a “ready” flag.

- **HTTP API Server:**  
  A built-in server exposes an endpoint (e.g. `/services`) that returns a JSON array of service statuses. This API lets clients quickly determine the availability of critical services.

---

## Demo

![Quick demo of maintenance-window-operator](docs/demo.gif)

---

## Quickstart

### Prerequisites

- A running Kubernetes cluster (v1.20+ recommended)
- [Helm 3](https://helm.sh/) installed

### Installation

1. **Install CRDs:**

   Ensure the Custom Resource Definitions are installed first:
   ```bash
   kubectl apply -k config/crd
   ```

2. **Deploy via Helm:**

   Install the operator using the provided Helm chart:
   ```bash
   helm install maintenance-window-manager ./charts/maintenance-window-manager \
     --namespace maintenance-window-manager-system --create-namespace
   ```

---

## Usage

Create a `ServiceChecker` resource to specify which services to monitor. For example:

```yaml
apiVersion: maintenance.mamrezb.com/v1alpha1
kind: ServiceChecker
metadata:
  name: example-checker
spec:
  services:
    - name: my-service
      namespace: default
      critical: true
```

Apply the resource:

```bash
kubectl apply -f example-servicechecker.yaml
```

Query the health status by accessing the HTTP API (adjust the service URL as needed):

```bash
curl -k https://<operator-service>.<namespace>.svc.cluster.local:8443/services
```

A sample response might look like:

```json
[
  {
    "name": "my-service",
    "namespace": "default",
    "ready": true,
    "critical": true
  }
]
```

---

## Development

To build and test the operator locally:

```bash
# Build the Docker image
make docker-build

# If using a Kind cluster, load the image
make docker-load

# Deploy the operator locally
make deploy

# Run tests
make test
```

---

## Contributing

Contributions are welcome. Please open GitHub issues or submit pull requests for bug fixes, improvements, or new features. Ensure that new code is covered by tests.

---

## Code of Conduct

This project follows a **strict code of conduct** to ensure a positive environment.  
Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

---

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
