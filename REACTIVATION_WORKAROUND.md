# PRStack Reactivation Workaround

## Current Issue

When setting `spec.active: true` on an expired PR stack, the operator may immediately set it back to `false` due to the expiration check running before the `LastActiveAt` timestamp is updated.

## Workaround: Manual LastActiveAt Update

Instead of just setting `active: true`, patch both fields at once:

```bash
# Set active=true AND update lastActiveAt in one operation
kubectl patch prstack pr-33 --type=merge -p "{
  \"spec\": {
    \"active\": true
  },
  \"status\": {
    \"lastActiveAt\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"
  }
}"
```

##Alternative: Delete and Recreate

If the stack is very old and you want a fresh start:

```bash
# Backup the spec first
kubectl get prstack pr-33 -o yaml > pr-33-backup.yaml

# Delete
kubectl delete prstack pr-33

# Recreate (edit the YAML to remove status, resourceVersion, etc.)
kubectl apply -f pr-33-backup.yaml
```

## Rollout Feature Works Independently

The deployment rollout feature works regardless of the reactivation issue:

```bash
# Once stack is active, trigger rollout
kubectl patch prstack pr-33 --type=merge -p "{\"spec\":{\"deployedAt\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}}"
```

## Status

This is a known issue being worked on. The core problem is a race condition between:
1. Patching `spec.active=true`
2. Reconcile loop checking expiration (which sets it back to `false`)
3. Updating `status.lastActiveAt`

A proper fix requires ensuring the `lastActiveAt` update happens atomically with the expiration check.

