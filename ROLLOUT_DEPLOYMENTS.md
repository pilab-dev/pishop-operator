# Deployment Rollout Feature

## Overview

The PRStack CRD now supports triggering a rollout of all deployments by updating the `deployedAt` timestamp in the spec. This is useful when you want to:
- Force all pods to restart with the latest configuration
- Apply updated image tags across all services
- Troubleshoot issues by restarting all services
- Ensure all services pick up new secrets or configmaps

## How It Works

1. **Set `spec.deployedAt`**: Update the `deployedAt` field with the current timestamp
2. **Controller Detects Change**: The operator compares `spec.deployedAt` with `status.lastDeployedAt`
3. **Rollout Triggered**: If different, all deployments in the namespace are rolled out
4. **Status Updated**: The `status.lastDeployedAt` is updated to match `spec.deployedAt`

The rollout is performed by updating the `kubectl.kubernetes.io/restartedAt` annotation on each deployment's pod template, which triggers Kubernetes to perform a rolling restart.

## Usage Examples

### Trigger a Rollout

```bash
# Get current timestamp in RFC3339 format
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Update the PRStack to trigger rollout
kubectl patch prstack pr-33 --type=merge -p "{\"spec\":{\"deployedAt\":\"$TIMESTAMP\"}}"
```

### Using kubectl edit

```bash
kubectl edit prstack pr-33
```

Then add or update the `deployedAt` field:

```yaml
spec:
  active: true
  deployedAt: "2025-10-02T14:30:00Z"  # Update this timestamp
  environment: pr-33
  imageTag: pr-33
  prNumber: "33"
  # ... rest of spec
```

### Using a YAML patch file

```bash
# Create patch file
cat <<EOF > rollout-patch.yaml
spec:
  deployedAt: "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
EOF

# Apply the patch
kubectl patch prstack pr-33 --patch-file rollout-patch.yaml
```

### Programmatic Rollout (from GitHub Actions)

```yaml
- name: Trigger Deployment Rollout
  run: |
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    kubectl patch prstack pr-${{ env.PR_NUMBER }} \
      --type=merge \
      -p "{\"spec\":{\"deployedAt\":\"$TIMESTAMP\"}}"
    
    echo "Rollout triggered at $TIMESTAMP"
```

## Verification

### Check Rollout Status

```bash
# View the deployment status
kubectl get deployments -n pr-33-shop-pilab-hu -o wide

# Check the restartedAt annotation
kubectl get deployment -n pr-33-shop-pilab-hu product-service \
  -o jsonpath='{.spec.template.metadata.annotations.kubectl\.kubernetes\.io/restartedAt}'

# View PRStack status
kubectl get prstack pr-33 -o yaml | grep -A 2 deployedAt
```

### Monitor Rollout Progress

```bash
# Watch all deployments
watch kubectl get deployments -n pr-33-shop-pilab-hu

# Watch pods during rollout
kubectl get pods -n pr-33-shop-pilab-hu -w

# Check rollout status for specific deployment
kubectl rollout status deployment/product-service -n pr-33-shop-pilab-hu
```

### View Operator Events

```bash
# Check events related to the PRStack
kubectl describe prstack pr-33 | grep -A 10 Events

# View operator logs
kubectl logs -n pishop-operator-system -l control-plane=controller-manager -f
```

## Expected Events

When a rollout is triggered, you should see these events:

```
Normal  RolloutTriggered  PR #33 deployments rolled out successfully
```

If there's an issue:

```
Warning RolloutFailed  Failed to rollout deployments: <error details>
```

## Integration with Image Updates

This feature works well with image tag updates:

```bash
# Update both image tag and trigger rollout in one command
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
kubectl patch prstack pr-33 --type=merge -p "{
  \"spec\": {
    \"imageTag\": \"pr-33-${{ github.sha }}\",
    \"deployedAt\": \"$TIMESTAMP\"
  }
}"
```

## Status Fields

The PRStack status includes:

- **`status.lastDeployedAt`**: Timestamp of the last successful rollout
- **`status.phase`**: Current phase (should remain in current phase during rollout)
- **`status.message`**: Error message if rollout fails

## Troubleshooting

### Rollout Not Triggering

```bash
# Check if deployedAt is set
kubectl get prstack pr-33 -o jsonpath='{.spec.deployedAt}'

# Check if it matches lastDeployedAt
kubectl get prstack pr-33 -o jsonpath='{.status.lastDeployedAt}'

# If they match, update deployedAt to a new timestamp
kubectl patch prstack pr-33 --type=merge -p "{\"spec\":{\"deployedAt\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}}"
```

### Deployments Not Restarting

```bash
# Check deployment annotations
kubectl get deployment -n pr-33-shop-pilab-hu -o yaml | grep -A 5 annotations

# Manually check deployment rollout status
kubectl rollout restart deployment/product-service -n pr-33-shop-pilab-hu
```

### Operator Not Responding

```bash
# Check operator logs
kubectl logs -n pishop-operator-system -l control-plane=controller-manager --tail=100

# Restart operator
kubectl rollout restart deployment/pishop-operator-controller-manager -n pishop-operator-system
```

## Limitations

- Only triggers rollout for existing deployments in the namespace
- Doesn't create new deployments or update image tags (use separate spec fields)
- Stack must be in `active: true` state for rollout to work
- Rollout happens asynchronously; check deployment status to confirm completion

## Best Practices

1. **Always use RFC3339 format**: `YYYY-MM-DDTHH:MM:SSZ`
2. **Wait for previous rollout to complete** before triggering another
3. **Monitor rollout progress** to ensure all pods are healthy
4. **Use with image tag updates** to ensure pods pull latest images
5. **Set `imagePullPolicy: Always`** in deployments for reliable updates

