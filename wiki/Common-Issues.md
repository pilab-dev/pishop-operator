# Common Issues and Troubleshooting

This guide covers common issues you may encounter when using the PiShop Operator and provides solutions to resolve them.

## Installation Issues

### CRD Installation Fails

**Symptoms:**
- `make install` command fails
- Error: "unable to recognize no matches for kind"

**Solutions:**
```bash
# Check cluster permissions
kubectl auth can-i create crd

# Try manual installation
kubectl apply -f config/crd/bases/shop.pilab.hu_prstacks.yaml

# Check if CRD already exists
kubectl get crd prstacks.shop.pilab.hu

# Delete and reinstall if corrupted
kubectl delete crd prstacks.shop.pilab.hu
kubectl apply -f config/crd/bases/shop.pilab.hu_prstacks.yaml
```

### Secret Creation Fails

**Symptoms:**
- `kubectl create secret` command fails
- Error: "namespace does not exist"

**Solutions:**
```bash
# Check if namespace exists
kubectl get namespace pishop-operator-system

# Create namespace if needed
kubectl create namespace pishop-operator-system

# Verify secret creation
kubectl get secrets -n pishop-operator-system
```

### Image Pull Errors

**Symptoms:**
- Pods stuck in `ImagePullBackOff` state
- Error: "failed to pull image"

**Solutions:**
```bash
# Check image pull secrets
kubectl get secrets -n pishop-operator-system

# Verify GitHub credentials
kubectl get secret github-registry-credentials -n pishop-operator-system -o yaml

# Check if image exists
docker pull ghcr.io/pilab-dev/product-service:pr-123

# Manually create image pull secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=your-github-username \
  --docker-password=ghp_your_token \
  --docker-email=your-email@example.com \
  --namespace=pishop-operator-system
```

## PRStack Issues

### PRStack Stuck in Pending Phase

**Symptoms:**
- PRStack status shows `phase: Pending`
- No namespace or resources created

**Solutions:**
```bash
# Check operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator

# Check PRStack events
kubectl describe prstack pr-123 | grep -A 10 Events

# Check if operator is running
kubectl get pods -n pishop-operator-system

# Restart operator if needed
kubectl rollout restart deployment/pishop-operator -n pishop-operator-system
```

### Services Not Starting

**Symptoms:**
- Deployments created but pods not running
- Pods in `CrashLoopBackOff` state

**Solutions:**
```bash
# Check deployment status
kubectl get deployments -n pr-123-shop-pilab-hu

# Check pod status and logs
kubectl get pods -n pr-123-shop-pilab-hu
kubectl logs -n pr-123-shop-pilab-hu deployment/product-service

# Check pod events
kubectl describe pod -n pr-123-shop-pilab-hu -l app=product-service

# Check resource limits
kubectl describe pod -n pr-123-shop-pilab-hu -l app=product-service | grep -A 5 "Limits\|Requests"
```

### Database Connection Issues

**Symptoms:**
- Services can't connect to MongoDB
- Error: "authentication failed"

**Solutions:**
```bash
# Check MongoDB credentials
kubectl get secret mongodb-credentials -n pishop-operator-system

# Test MongoDB connectivity
kubectl run mongodb-test --image=mongo:latest --rm -it --restart=Never -- \
  mongo "mongodb://admin:password@mongodb:27017/admin"

# Check MongoDB user creation
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i mongo

# Verify MongoDB is accessible from cluster
kubectl run network-test --image=busybox --rm -it --restart=Never -- \
  wget -qO- mongodb:27017
```

### NATS Connection Issues

**Symptoms:**
- Services can't connect to NATS
- Error: "connection refused"

**Solutions:**
```bash
# Check NATS server status
kubectl get pods -n pishop-base -l app=nats

# Test NATS connectivity
kubectl run nats-test --image=nats:latest --rm -it --restart=Never -- \
  nats pub test "Hello World"

# Check NATS configuration
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i nats
```

### Redis Connection Issues

**Symptoms:**
- Services can't connect to Redis
- Error: "connection refused"

**Solutions:**
```bash
# Check Redis server status
kubectl get pods -n pishop-base -l app=redis

# Test Redis connectivity
kubectl run redis-test --image=redis:latest --rm -it --restart=Never -- \
  redis-cli -h redis ping

# Check Redis configuration
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i redis
```

## Network Issues

### Custom Domain Not Working

**Symptoms:**
- Custom domain returns 404 or connection refused
- DNS resolution fails

**Solutions:**
```bash
# Check DNS resolution
nslookup magicshop.hu
dig magicshop.hu

# Check ingress configuration
kubectl get ingress -n pr-123-shop-pilab-hu -o yaml

# Check ingress controller
kubectl get pods -n ingress-nginx

# Check TLS secret
kubectl get secret magicshop-tls -n default

# Verify ingress controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
```

### TLS Certificate Issues

**Symptoms:**
- HTTPS connection fails
- Certificate errors in browser

**Solutions:**
```bash
# Check TLS secret
kubectl get secret magicshop-tls -n default -o yaml

# Verify certificate validity
kubectl get secret magicshop-tls -n default -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -text -noout

# Check cert-manager (if using)
kubectl get certificates -n default
kubectl describe certificate magicshop-tls -n default

# Check certificate issuer
kubectl get clusterissuer
```

## Resource Issues

### Out of Memory Errors

**Symptoms:**
- Pods killed with OOMKilled status
- Memory usage exceeds limits

**Solutions:**
```bash
# Check resource usage
kubectl top pods -n pr-123-shop-pilab-hu

# Increase memory limits
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "resourceLimits": {
      "memoryLimit": "2Gi"
    }
  }
}'

# Check node resources
kubectl describe nodes
```

### CPU Throttling

**Symptoms:**
- Services running slowly
- High CPU throttling metrics

**Solutions:**
```bash
# Check CPU usage
kubectl top pods -n pr-123-shop-pilab-hu

# Increase CPU limits
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "resourceLimits": {
      "cpuLimit": "1000m"
    }
  }
}'

# Check node capacity
kubectl describe nodes
```

### Storage Issues

**Symptoms:**
- PVC creation fails
- Database storage full

**Solutions:**
```bash
# Check storage classes
kubectl get storageclass

# Check PVC status
kubectl get pvc -n pr-123-shop-pilab-hu

# Increase storage limits
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "resourceLimits": {
      "storageLimit": "50Gi"
    }
  }
}'

# Check available storage
kubectl describe nodes | grep -A 5 "Allocated resources"
```

## Backup and Restore Issues

### Backup Jobs Failing

**Symptoms:**
- Backup jobs in Failed state
- No backups being created

**Solutions:**
```bash
# Check backup job status
kubectl get jobs -n pr-123-shop-pilab-hu | grep backup

# Check backup job logs
kubectl logs -n pr-123-shop-pilab-hu job/backup-pr-123-20240115

# Check backup configuration
kubectl get prstack pr-123 -o yaml | grep -A 10 backupConfig

# Verify storage class
kubectl get storageclass fast-ssd
```

### Restore Operations Failing

**Symptoms:**
- Restore jobs failing
- Data not restored correctly

**Solutions:**
```bash
# Check restore job status
kubectl get jobs -n pr-123-shop-pilab-hu | grep restore

# Check restore job logs
kubectl logs -n pr-123-shop-pilab-hu job/restore-pr-123-20240115

# Verify backup exists
kubectl get pvc -n pr-123-shop-pilab-hu | grep backup

# Check MongoDB connectivity during restore
kubectl logs -n pr-123-shop-pilab-hu job/restore-pr-123-20240115 | grep -i mongo
```

## Operator Issues

### Operator Not Responding

**Symptoms:**
- PRStack changes not being processed
- No events or status updates

**Solutions:**
```bash
# Check operator pod status
kubectl get pods -n pishop-operator-system

# Check operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator --tail=100

# Restart operator
kubectl rollout restart deployment/pishop-operator -n pishop-operator-system

# Check leader election
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i leader
```

### High Memory Usage

**Symptoms:**
- Operator pod using excessive memory
- Potential memory leaks

**Solutions:**
```bash
# Check operator resource usage
kubectl top pods -n pishop-operator-system

# Check operator logs for memory issues
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i memory

# Restart operator
kubectl rollout restart deployment/pishop-operator -n pishop-operator-system

# Check for resource limits
kubectl describe pod -n pishop-operator-system -l control-plane=controller-manager
```

### Reconcile Loops

**Symptoms:**
- Continuous reconcile operations
- High CPU usage on operator

**Solutions:**
```bash
# Check reconcile frequency
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i reconcile

# Check for error conditions
kubectl get prstacks --all-namespaces -o yaml | grep -A 5 -B 5 "condition"

# Check operator metrics
kubectl port-forward -n pishop-operator-system deployment/pishop-operator 8080:8080
curl http://localhost:8080/metrics | grep reconcile
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Enable debug logging
kubectl patch deployment pishop-operator -n pishop-operator-system -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "manager",
          "env": [{
            "name": "LOG_LEVEL",
            "value": "debug"
          }]
        }]
      }
    }
  }
}'

# View debug logs
kubectl logs -n pishop-operator-system deployment/pishop-operator -f
```

### Check Resource Status

```bash
# Check all PRStacks
kubectl get prstacks --all-namespaces

# Check operator status
kubectl get pods -n pishop-operator-system

# Check external service connectivity
kubectl run connectivity-test --image=busybox --rm -it --restart=Never -- \
  wget -qO- mongodb:27017 && \
  wget -qO- nats:4222 && \
  wget -qO- redis:6379
```

### Monitor Events

```bash
# Watch all events
kubectl get events --all-namespaces -w

# Watch PRStack events
kubectl get events --field-selector involvedObject.name=pr-123 -w

# Check operator events
kubectl get events -n pishop-operator-system -w
```

## Getting Help

### Log Collection

When reporting issues, collect the following information:

```bash
# Operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator > operator-logs.txt

# PRStack status
kubectl get prstack pr-123 -o yaml > prstack-status.yaml

# Cluster events
kubectl get events --all-namespaces > cluster-events.txt

# Resource status
kubectl get all -n pr-123-shop-pilab-hu > pr-resources.txt
```

### Support Channels

- **GitHub Issues**: [Report bugs and request features](https://github.com/pilab-dev/pishop-operator/issues)
- **Discussions**: [Community discussions](https://github.com/pilab-dev/pishop-operator/discussions)
- **Email**: support@pilab.hu
- **Discord**: [PiLab Community](https://discord.gg/pilab)

### Before Reporting Issues

1. Check this troubleshooting guide
2. Search existing GitHub issues
3. Enable debug logging and collect logs
4. Include cluster information (Kubernetes version, operator version)
5. Provide minimal reproduction steps

## Prevention

### Best Practices

1. **Regular Monitoring**: Set up monitoring and alerting
2. **Resource Planning**: Plan resource requirements carefully
3. **Backup Strategy**: Implement regular backup procedures
4. **Testing**: Test in non-production environments first
5. **Documentation**: Keep configuration documentation up to date

### Health Checks

```bash
# Regular health check script
#!/bin/bash
echo "Checking PiShop Operator health..."

# Check operator status
kubectl get pods -n pishop-operator-system

# Check PRStack status
kubectl get prstacks --all-namespaces

# Check resource usage
kubectl top pods -n pishop-operator-system

# Check external connectivity
kubectl run health-check --image=busybox --rm --restart=Never -- \
  wget -qO- mongodb:27017 && echo "MongoDB: OK" || echo "MongoDB: FAIL"
```

## Next Steps

After resolving issues:

1. [Set up monitoring](Monitoring-Observability) to prevent future issues
2. [Review configuration](Configuration) for optimization opportunities
3. [Implement backup strategies](Backup-Restore) for data protection
4. [Document solutions](https://github.com/pilab-dev/pishop-operator/wiki) for future reference
