# 🚀 PiShop Operator

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24.0-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.29.0-blue.svg)](https://kubernetes.io/)
[![Docker](https://img.shields.io/badge/Docker-GHCR-blue.svg)](https://ghcr.io/pilab-dev/pishop-operator)
[![Build Status](https://github.com/pilab-dev/pishop-operator/workflows/Build%20Operator/badge.svg)](https://github.com/pilab-dev/pishop-operator/actions)
[![Go Report Card](https://goreportcard.com/badge/go.pilab.hu/shop/pishop-provisioner)](https://goreportcard.com/report/go.pilab.hu/shop/pishop-provisioner)

> A sophisticated Kubernetes operator for managing PR-based environments for PiShop microservices with automated provisioning, isolation, and lifecycle management.

## ✨ Features

### 🔧 Core Capabilities
- **Automated PR Environment Provisioning**: Spin up complete isolated environments for each pull request
- **Database Management**: Automated MongoDB database and user creation per PR
- **Message Queue Isolation**: NATS subject management for complete PR isolation
- **Cache Isolation**: Redis keyspace isolation for each PR environment
- **Service Orchestration**: Automated deployment and management of microservices
- **Resource Management**: Configurable CPU, memory, and storage limits
- **Backup & Restore**: Automated backup scheduling and restore capabilities

### 🎯 Advanced Features
- **GitHub Integration**: Seamless integration with GitHub Container Registry (GHCR)
- **Automatic Cleanup**: Intelligent expiration and cleanup of PR environments
- **Health Monitoring**: Comprehensive status tracking and health checks
- **Rolling Updates**: Zero-downtime deployments with configurable strategies
- **Environment Templates**: Reusable configuration templates for different environments

## 🏗️ Architecture

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "PR Namespace (pr-123-shop-pilab-hu)"
            PRStack[PRStack CRD]
            MongoDB[(MongoDB<br/>Database)]
            NATS[NATS<br/>Message Queue]
            Redis[(Redis<br/>Cache)]
            Services[Microservices<br/>Deployments]
        end
        
        subgraph "PiShop Operator System"
            Operator[PiShop Operator]
            Controller[PRStack Controller]
        end
        
        subgraph "External Services"
            GHCR[GitHub Container<br/>Registry]
            GitHub[GitHub API]
        end
    end
    
    PRStack --> Controller
    Controller --> Operator
    Operator --> MongoDB
    Operator --> NATS
    Operator --> Redis
    Operator --> Services
    Operator --> GHCR
    Operator --> GitHub
```

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster (v1.29+)
- kubectl configured
- MongoDB instance
- NATS server
- Redis instance
- GitHub Personal Access Token with `packages:read` scope

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/pilab-dev/pishop-operator.git
   cd pishop-operator
   ```

2. **Install CRDs**
   ```bash
   make install
   ```

3. **Configure secrets**
```bash
   # MongoDB credentials
kubectl create secret generic mongodb-credentials \
  --namespace=pishop-operator-system \
  --from-literal=uri=mongodb://admin:password@mongodb.pishop-base.svc.cluster.local:27017 \
  --from-literal=username=admin \
  --from-literal=password=password

   # GitHub credentials
kubectl create secret generic github-registry-credentials \
  --namespace=pishop-operator-system \
  --from-literal=username=your-github-username \
  --from-literal=token=ghp_your_github_token \
  --from-literal=email=your-email@example.com
```

4. **Deploy the operator**
```bash
   make deploy
   ```

## 📖 Usage

### Creating a PR Stack

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-123
  namespace: pr-123-shop-pilab-hu
spec:
  prNumber: "123"
  imageTag: "pr-123-abc123"  # optional, defaults to pr-{prNumber}
  active: true
  environment: "development"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
  resourceLimits:
    cpuLimit: "500m"
    memoryLimit: "1Gi"
    storageLimit: "10Gi"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"  # Daily at 2 AM
    retentionDays: 7
    storageSize: "5Gi"
```

### Advanced Configuration

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-456
spec:
  prNumber: "456"
  active: true
  environment: "staging"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
    - notification-service
  # Custom connection details (optional)
  mongoURI: "mongodb://custom-mongo:27017"
  natsURL: "nats://custom-nats:4222"
  redisURL: "redis://custom-redis:6379"
  resourceLimits:
    cpuLimit: "1000m"
    memoryLimit: "2Gi"
    storageLimit: "20Gi"
  backupConfig:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    retentionDays: 14
    storageClass: "fast-ssd"
    storageSize: "10Gi"
```

## 🔧 Configuration

### Environment Variables

The operator reads configuration from the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `MONGO_URI` | MongoDB connection URI | Yes |
| `MONGO_USERNAME` | MongoDB admin username | Yes |
| `MONGO_PASSWORD` | MongoDB admin password | Yes |
| `GITHUB_USERNAME` | GitHub username for GHCR | Yes |
| `GITHUB_TOKEN` | GitHub token for GHCR | Yes |
| `GITHUB_EMAIL` | GitHub email (optional) | No |

### Resource Limits

Configure resource constraints for PR environments:

```yaml
resourceLimits:
  cpuLimit: "500m"        # CPU limit per service
  memoryLimit: "1Gi"      # Memory limit per service
  storageLimit: "10Gi"    # Storage limit for databases
```

### Backup Configuration

Enable automated backups with configurable schedules:

```yaml
backupConfig:
  enabled: true                    # Enable backups
  schedule: "0 2 * * *"           # Cron schedule (daily at 2 AM)
  retentionDays: 7                 # Keep backups for 7 days
  storageClass: "standard"         # Storage class for backup PVCs
  storageSize: "5Gi"               # Size of backup storage
```

## 🎛️ Management Commands

### Development

```bash
# Build the operator
make build

# Run locally
make run

# Run tests
make test

# Generate manifests
make manifests

# Format code
make fmt

# Lint code
make vet
```

### Deployment

```bash
# Deploy to cluster
make deploy

# Check deployment status
make status

# View logs
make logs

# Restart operator
make restart

# Undeploy
make undeploy
```

### Monitoring

```bash
# Check PRStacks
kubectl get prstacks

# Describe a PRStack
kubectl describe prstack pr-123

# Check operator logs
kubectl logs -f deployment/pishop-operator -n pishop-operator-system

# Check all resources in PR namespace
kubectl get all -n pr-123-shop-pilab-hu
```

## 🔍 Troubleshooting

### Common Issues

#### Image Pull Errors

If you see `ImagePullBackOff` errors:

1. **Check registry secret**
   ```bash
   kubectl get secret ghcr-secret -n pr-123-shop-pilab-hu
   ```

2. **Verify GitHub credentials**
   ```bash
   kubectl get secret github-registry-credentials -n pishop-operator-system
   ```

3. **Check operator logs**
   ```bash
   kubectl logs -n pishop-operator-system deployment/pishop-operator
   ```

#### Database Connection Issues

1. **Verify MongoDB credentials**
   ```bash
   kubectl get secret mongodb-credentials -n pishop-operator-system
   ```

2. **Check database status**
   ```bash
   kubectl describe prstack pr-123 | grep -A 10 "MongoDB"
   ```

#### Service Deployment Issues

1. **Check service status**
   ```bash
   kubectl get deployments -n pr-123-shop-pilab-hu
   kubectl describe deployment product-service -n pr-123-shop-pilab-hu
   ```

2. **Check pod logs**
   ```bash
   kubectl logs -f deployment/product-service -n pr-123-shop-pilab-hu
   ```

### Debug Mode

Enable debug logging by setting the log level:

```bash
kubectl patch deployment pishop-operator -n pishop-operator-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'
```

## 📊 Status and Monitoring

### PRStack Status

The operator provides comprehensive status information:

```yaml
status:
  phase: "Running"
  message: "All services deployed successfully"
  createdAt: "2024-01-15T10:30:00Z"
  lastActiveAt: "2024-01-15T14:22:00Z"
  lastDeployedAt: "2024-01-15T10:35:00Z"
  mongodb:
    user: "pr-123-user"
    connectionString: "mongodb://pr-123-user:password@mongodb:27017/pr-123-db"
    databases: ["pr-123-product", "pr-123-cart", "pr-123-order"]
  nats:
    subjectPrefix: "pr-123"
    connectionString: "nats://nats:4222"
  redis:
    keyPrefix: "pr-123:"
    connectionString: "redis://redis:6379/0"
  services:
    - name: "product-service"
      status: "Running"
      url: "http://product-service.pr-123-shop-pilab-hu.svc.cluster.local:8080"
    - name: "cart-service"
      status: "Running"
      url: "http://cart-service.pr-123-shop-pilab-hu.svc.cluster.local:8080"
  backup:
    lastBackupTime: "2024-01-15T02:00:00Z"
    lastBackupName: "backup-pr-123-20240115"
    backupCount: 3
    lastBackupSize: "2.1Gi"
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Run tests**
   ```bash
   make test
   ```
5. **Commit your changes**
   ```bash
   git commit -m 'Add amazing feature'
   ```
6. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Kubernetes](https://kubernetes.io/) for the amazing orchestration platform
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) for the operator framework
- [MongoDB](https://www.mongodb.com/) for the database
- [NATS](https://nats.io/) for the message queue
- [Redis](https://redis.io/) for the cache
- [GitHub](https://github.com/) for the container registry

## 📞 Support

- 📧 Email: support@pilab.hu
- 💬 Discord: [PiLab Community](https://discord.gg/pilab)
- 📖 Documentation: [docs.pilab.hu](https://docs.pilab.hu)
- 🐛 Issues: [GitHub Issues](https://github.com/pilab-dev/pishop-operator/issues)

---

<div align="center">
  <strong>Built with ❤️ by the PiLab Team</strong>
</div>