# Installation Guide

This guide walks you through installing the PiShop Operator on your Kubernetes cluster.

## Prerequisites

### System Requirements
- Kubernetes cluster (v1.29+)
- kubectl configured and connected to your cluster
- Helm (optional, for Helm-based installation)

### External Dependencies
- **MongoDB**: MongoDB instance accessible from the cluster
- **NATS**: NATS server for message queuing
- **Redis**: Redis instance for caching
- **GitHub Access**: GitHub Personal Access Token with `packages:read` scope

### Cluster Requirements
- Cluster admin privileges for CRD installation
- Sufficient resources for operator deployment
- Network connectivity to external services

## Installation Methods

### Method 1: Direct kubectl Installation (Recommended)

#### Step 1: Clone the Repository
```bash
git clone https://github.com/pilab-dev/pishop-operator.git
cd pishop-operator
```

#### Step 2: Install CRDs
```bash
make install
```

This command installs the PRStack Custom Resource Definition and RBAC resources.

#### Step 3: Configure Secrets

**MongoDB Credentials:**
```bash
kubectl create secret generic mongodb-credentials \
  --namespace=pishop-operator-system \
  --from-literal=uri=mongodb://admin:password@mongodb.pishop-base.svc.cluster.local:27017 \
  --from-literal=username=admin \
  --from-literal=password=password
```

**GitHub Registry Credentials:**
```bash
kubectl create secret generic github-registry-credentials \
  --namespace=pishop-operator-system \
  --from-literal=username=your-github-username \
  --from-literal=token=ghp_your_github_token \
  --from-literal=email=your-email@example.com
```

#### Step 4: Deploy the Operator
```bash
make deploy
```

#### Step 5: Verify Installation
```bash
# Check operator deployment
kubectl get pods -n pishop-operator-system

# Check CRD installation
kubectl get crd prstacks.shop.pilab.hu

# Check operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator
```

### Method 2: Helm Installation

#### Step 1: Add the Helm Repository
```bash
helm repo add pishop-operator https://pilab-dev.github.io/pishop-operator
helm repo update
```

#### Step 2: Create Values File
```yaml
# values.yaml
operator:
  image:
    repository: ghcr.io/pilab-dev/pishop-operator
    tag: latest
  
  config:
    baseDomain: "shop.pilab.hu"
    mongoURI: "mongodb://admin:password@mongodb.pishop-base.svc.cluster.local:27017"
    mongoUsername: "admin"
    mongoPassword: "password"

secrets:
  mongodb:
    existingSecret: "mongodb-credentials"
  github:
    existingSecret: "github-registry-credentials"
```

#### Step 3: Install with Helm
```bash
helm install pishop-operator pishop-operator/pishop-operator \
  --namespace pishop-operator-system \
  --create-namespace \
  --values values.yaml
```

### Method 3: Operator Lifecycle Manager (OLM)

#### Step 1: Install OLM
```bash
curl -sL https://github.com/operator-framework/operator-sdk/releases/download/v1.34.0/operator-sdk_linux_amd64 -o operator-sdk
chmod +x operator-sdk
sudo mv operator-sdk /usr/local/bin/

operator-sdk olm install
```

#### Step 2: Create OperatorGroup
```yaml
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: pishop-operator-group
  namespace: pishop-operator-system
spec:
  targetNamespaces:
  - pishop-operator-system
```

#### Step 3: Install via OLM
```bash
kubectl apply -f https://raw.githubusercontent.com/pilab-dev/pishop-operator/main/config/olm/catalog-source.yaml
kubectl apply -f https://raw.githubusercontent.com/pilab-dev/pishop-operator/main/config/olm/subscription.yaml
```

## Configuration

### Environment Variables

Configure the operator using environment variables or ConfigMaps:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: BASE_DOMAIN
          value: "your-domain.com"
        - name: MONGO_URI
          valueFrom:
            secretKeyRef:
              name: mongodb-credentials
              key: uri
        - name: GITHUB_USERNAME
          valueFrom:
            secretKeyRef:
              name: github-registry-credentials
              key: username
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: github-registry-credentials
              key: token
```

### Command Line Flags

The operator supports command-line flags that override environment variables:

```bash
./pishop-operator \
  --mongo-uri="mongodb://localhost:27017" \
  --base-domain="example.com" \
  --github-username="myuser" \
  --github-token="ghp_xxx"
```

Available flags:
- `--mongo-uri`: MongoDB connection URI (required)
- `--mongo-username`: MongoDB admin username
- `--mongo-password`: MongoDB admin password
- `--base-domain`: Base domain for default PR domains
- `--github-username`: GitHub username for container registry
- `--github-token`: GitHub token for container registry
- `--github-email`: GitHub email for container registry
- `--metrics-bind-address`: Metrics server bind address (default: ":8080")
- `--health-probe-bind-address`: Health probe bind address (default: ":8081")
- `--leader-elect`: Enable leader election (default: false)

## Post-Installation Verification

### 1. Check Operator Status
```bash
kubectl get pods -n pishop-operator-system
kubectl describe pod -n pishop-operator-system -l control-plane=controller-manager
```

### 2. Verify CRD Installation
```bash
kubectl get crd prstacks.shop.pilab.hu -o yaml
```

### 3. Test PRStack Creation
```bash
# Create a test PRStack
cat <<EOF | kubectl apply -f -
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: test-stack
spec:
  prNumber: "test"
  active: true
  services:
    - product-service
EOF

# Check the status
kubectl get prstack test-stack
kubectl describe prstack test-stack

# Clean up
kubectl delete prstack test-stack
```

### 4. Verify External Service Connectivity
```bash
# Check MongoDB connectivity
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i mongo

# Check GitHub connectivity
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i github

# Check NATS connectivity
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i nats
```

## Troubleshooting Installation

### Common Issues

#### 1. CRD Installation Fails
```bash
# Check cluster permissions
kubectl auth can-i create crd

# Try manual installation
kubectl apply -f config/crd/bases/shop.pilab.hu_prstacks.yaml
```

#### 2. Secret Creation Fails
```bash
# Check namespace exists
kubectl get namespace pishop-operator-system

# Create namespace if needed
kubectl create namespace pishop-operator-system
```

#### 3. Image Pull Errors
```bash
# Check image pull secrets
kubectl get secrets -n pishop-operator-system

# Verify GitHub credentials
kubectl get secret github-registry-credentials -n pishop-operator-system -o yaml
```

#### 4. MongoDB Connection Issues
```bash
# Test MongoDB connectivity from cluster
kubectl run mongodb-test --image=mongo:latest --rm -it --restart=Never -- \
  mongo "mongodb://admin:password@mongodb.pishop-base.svc.cluster.local:27017/admin"
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
kubectl patch deployment pishop-operator -n pishop-operator-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'
```

## Upgrading

### Upgrade Procedure

1. **Backup current configuration:**
```bash
kubectl get prstacks --all-namespaces -o yaml > prstacks-backup.yaml
```

2. **Update to new version:**
```bash
git pull origin main
make install
make deploy
```

3. **Verify upgrade:**
```bash
kubectl get pods -n pishop-operator-system
kubectl logs -n pishop-operator-system deployment/pishop-operator
```

### Rollback Procedure

If issues occur after upgrade:

```bash
# Rollback to previous version
git checkout <previous-commit>
make install
make deploy
```

## Uninstallation

### Complete Removal

```bash
# Delete all PRStacks first
kubectl delete prstacks --all --all-namespaces

# Remove operator
make undeploy

# Remove CRDs (optional)
kubectl delete crd prstacks.shop.pilab.hu
```

### Helm Uninstallation

```bash
helm uninstall pishop-operator -n pishop-operator-system
```

## Next Steps

After successful installation:

1. [Configure the operator](Configuration) for your environment
2. [Create your first PRStack](Creating-PR-Stacks)
3. [Set up monitoring](Monitoring-Observability)
4. [Configure custom domains](Custom-Domains) if needed

## Support

If you encounter issues during installation:

- Check the [Troubleshooting Guide](Common-Issues)
- Review [GitHub Issues](https://github.com/pilab-dev/pishop-operator/issues)
- Join our [Discord Community](https://discord.gg/pilab)
