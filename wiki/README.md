# PiShop Operator Wiki

Welcome to the comprehensive documentation for the PiShop Operator! This wiki provides detailed guides for deploying, configuring, and managing PR-based environments for PiShop microservices.

## 📚 Documentation Structure

### Getting Started
- **[Home](Home)** - Overview and navigation
- **[Installation](Installation)** - Step-by-step installation guide
- **[Configuration](Configuration)** - Complete configuration reference
- **[Creating PR Stacks](Creating-PR-Stacks)** - How to create and manage PR environments

### Core Features
- **[Deployment Rollouts](Deployment-Rollouts)** - Triggering service rollouts
- **[API Reference](API-Reference)** - Complete PRStack CRD documentation
- **[Common Issues](Common-Issues)** - Troubleshooting guide

### Advanced Topics
- **[Advanced Features](Advanced-Features)** - Custom domains, monitoring, multi-tenant deployment

## 🚀 Quick Start

1. **Install the operator**: Follow the [Installation Guide](Installation)
2. **Configure secrets**: Set up MongoDB and GitHub credentials
3. **Create your first PRStack**: Use the [Creating PR Stacks](Creating-PR-Stacks) guide
4. **Monitor and troubleshoot**: Check [Common Issues](Common-Issues) if needed

## 📖 Key Concepts

### PRStack
A PRStack is a custom Kubernetes resource that represents a complete isolated environment for a pull request or tenant. It includes:
- Isolated namespace
- Dedicated MongoDB database and user
- NATS subject isolation
- Redis keyspace isolation
- Deployed microservices

### Lifecycle Phases
PRStacks go through several phases:
- `Pending` → `Provisioning` → `Running` → `Stopping` → `Stopped`
- Or `Failed` if errors occur

### Rollout Feature
The operator supports triggering rollouts by updating the `deployedAt` timestamp, useful for:
- Force pod restarts
- Apply configuration changes
- Troubleshoot issues

## 🔧 Common Operations

### Create a PRStack
```bash
kubectl apply -f - <<EOF
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-123
spec:
  prNumber: "123"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
EOF
```

### Trigger a Rollout
```bash
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
kubectl patch prstack pr-123 --type=merge -p "{\"spec\":{\"deployedAt\":\"$TIMESTAMP\"}}"
```

### Check Status
```bash
kubectl get prstacks
kubectl describe prstack pr-123
kubectl get all -n pr-123-shop-pilab-hu
```

## 📞 Support

- **GitHub Issues**: [Report bugs and request features](https://github.com/pilab-dev/pishop-operator/issues)
- **Discussions**: [Community discussions](https://github.com/pilab-dev/pishop-operator/discussions)
- **Email**: support@pilab.hu
- **Discord**: [PiLab Community](https://discord.gg/pilab)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/pilab-dev/pishop-operator/blob/main/LICENSE) file for details.

---

<div align="center">
  <strong>Built with ❤️ by the PiLab Team</strong>
</div>
