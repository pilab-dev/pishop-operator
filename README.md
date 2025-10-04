# PiShop Operator

Kubernetes operator for managing PR-based environments for PiShop microservices.

## Features

- Automated provisioning of MongoDB databases per PR
- NATS subject management for PR isolation
- Redis keyspace isolation
- Service deployment management
- Automated cleanup and expiration
- Backup and restore capabilities

## Configuration

### MongoDB Credentials

The operator requires MongoDB credentials to provision databases and users for PR environments.

#### Setting up MongoDB credentials

Update the `mongodb-credentials` secret:

```bash
kubectl create secret generic mongodb-credentials \
  --namespace=pishop-operator-system \
  --from-literal=uri=mongodb://admin:password@mongodb.pishop-base.svc.cluster.local:27017 \
  --from-literal=username=admin \
  --from-literal=password=password
```

Or edit `config/manager/mongodb-credentials-secret.yaml` and apply:

```bash
kubectl apply -f config/manager/mongodb-credentials-secret.yaml
```

### GitHub Container Registry Credentials

The operator needs GitHub credentials to pull images from GHCR and to create image pull secrets in PR namespaces.

#### Setting up GitHub credentials

1. Create a GitHub Personal Access Token with `packages:read` scope
2. Create the secret:

```bash
kubectl create secret generic github-registry-credentials \
  --namespace=pishop-operator-system \
  --from-literal=username=your-github-username \
  --from-literal=token=ghp_your_github_token \
  --from-literal=email=your-email@example.com
```

Or edit `config/manager/github-credentials-secret.yaml` and apply:

```bash
kubectl apply -f config/manager/github-credentials-secret.yaml
```

The operator will automatically:
- Read these credentials from environment variables
- Create `ghcr-secret` in each PR namespace
- Configure deployments to use this secret for pulling images

### Environment Variables

The operator reads configuration from the following environment variables (populated from secrets):

- `MONGO_URI`: MongoDB connection URI
- `MONGO_USERNAME`: MongoDB admin username
- `MONGO_PASSWORD`: MongoDB admin password
- `GITHUB_USERNAME`: GitHub username for container registry
- `GITHUB_TOKEN`: GitHub token for container registry
- `GITHUB_EMAIL`: GitHub email (optional)

## Deployment

### Using manager.yaml (production)

```bash
kubectl apply -f config/manager/manager.yaml
```

### Using manager-local.yaml (local testing)

```bash
kubectl apply -f config/manager/manager-local.yaml
```

### Using manager-simple.yaml (development)

```bash
kubectl apply -f config/manager/manager-simple.yaml
```

## Creating a PR Stack

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-123
spec:
  prNumber: "123"
  imageTag: "pr-123-abc123"  # optional, defaults to pr-{prNumber}
  active: true
  services:
    - product-service
    - cart-service
    - order-service
```

## How Registry Secrets Work

1. The operator deployment has access to GitHub credentials via the `github-registry-credentials` secret
2. When a PR namespace is created, the operator:
   - Reads the GitHub username and token from its environment
   - Creates a Docker config JSON with GHCR authentication
   - Creates a `ghcr-secret` of type `kubernetes.io/dockerconfigjson` in the PR namespace
3. All service deployments reference this `ghcr-secret` via `imagePullSecrets`
4. Kubernetes uses this secret to authenticate when pulling images from GHCR

## Troubleshooting

### Image pull errors

If you see `ImagePullBackOff` errors:

1. Check if the `ghcr-secret` exists in the PR namespace:
   ```bash
   kubectl get secret ghcr-secret -n pr-123-shop-pilab-hu
   ```

2. Verify the operator has GitHub credentials:
   ```bash
   kubectl get secret github-registry-credentials -n pishop-operator-system
   ```

3. Check operator logs:
   ```bash
   kubectl logs -n pishop-operator-system deployment/pishop-operator
   ```

### Missing credentials

If credentials are not configured, the operator will log:
```
GitHub credentials not configured, skipping registry secret creation
```

This is not an error if you're using public images, but required for private GHCR repositories.