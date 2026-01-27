# Cosign Integration for Drone-Docker

This document describes how to use the cosign container image signing feature in drone-docker.

## Overview

The drone-docker plugin now supports automatic container image signing using cosign after each successful push. This provides cryptographic verification that images haven't been tampered with.

## Environment Variables

The plugin accepts three cosign-related environment variables:

### `PLUGIN_COSIGN_PRIVATE_KEY` (Required for signing)
- **Description**: Private key for signing (PEM format content or file path)
- **Format**: Either PEM content or file path to private key
- **Usage**: Should be provided via secrets

### `PLUGIN_COSIGN_PASSWORD` (Optional)
- **Description**: Password for encrypted private keys
- **Usage**: Only needed if your private key is password-protected

### `PLUGIN_COSIGN_PARAMS` (Optional)
- **Description**: Additional cosign parameters
- **Examples**: 
  - `-a build_id=123` (add annotations)
  - `--tlog-upload=false` (disable transparency log)
  - `--rekor-url=https://custom-rekor.example.com` (custom rekor instance)

## Usage Examples

### 1. Basic Signing (Drone)

```yaml
kind: pipeline
type: docker
name: default

steps:
- name: docker
  image: plugins/docker
  settings:
    repo: myregistry/myapp
    tags: latest
    cosign_private_key:
      from_secret: cosign_private_key
    cosign_password:
      from_secret: cosign_password
```

### 2. Advanced Signing with Annotations (Drone)

```yaml
steps:
- name: docker
  image: plugins/docker
  settings:
    repo: myregistry/myapp
    tags: 
      - latest
      - ${DRONE_BUILD_NUMBER}
    cosign_private_key:
      from_secret: cosign_private_key
    cosign_params: "-a build_id=${DRONE_BUILD_NUMBER} -a commit_sha=${DRONE_COMMIT_SHA} -a branch=${DRONE_BRANCH}"
```

### 3. Harness CI/CD Usage

```yaml
- step:
    type: Plugin
    name: Build and Sign
    identifier: build_and_sign
    spec:
      connectorRef: account.harnessImage
      image: plugins/docker
      settings:
        repo: myregistry/myapp
        tags: <+pipeline.sequenceId>
        cosign_private_key: <+secrets.getValue("cosign_private_key")>
        cosign_password: <+secrets.getValue("cosign_password")>
        cosign_params: "-a harness_build=<+pipeline.sequenceId> -a harness_project=<+project.name>"
```

## Key Management

### Generating Cosign Keys

```bash
# Generate a new key pair
cosign generate-key-pair

# This creates:
# - cosign.key (private key) 
# - cosign.pub (public key)
```

### Storing Keys Securely
**Harness Secrets:**
1. Go to Project Settings → Secrets
2. Create new secret with type "File" for private key
3. Create new secret with type "Text" for password

## Security Features

### Automatic Validation
- ✅ **Private key format validation**: Ensures PEM format is correct
- ✅ **Password requirement detection**: Warns if encrypted key needs password
- ✅ **Keyless signing prevention**: Warns that OIDC keyless signing isn't supported

### Error Handling
- **Invalid private key**: `❌ Invalid private key format. Expected PEM format`
- **Missing password**: `🔐 Encrypted private key requires password. Set PLUGIN_COSIGN_PASSWORD`
- **Keyless signing**: `⚠️ WARNING: Keyless signing (OIDC) isn't supported yet in this plugin`

## Signing Behavior

### When Signing Occurs
- ✅ **After each successful push**: Images are signed immediately after push
- ✅ **Multiple tags**: Each tag gets signed individually
- ✅ **Push-only mode**: Works with existing images
- ✅ **Dry-run respect**: Skips signing in dry-run mode

### Image References
- **Preferred**: Signs by digest (e.g., `image@sha256:abc123...`) for security
- **Fallback**: Signs by tag if digest unavailable

### Authentication
- **Registry auth**: Automatically uses existing Docker registry credentials

## Verification

To verify a signed image:

```bash
# Verify with public key
cosign verify --key cosign.pub myregistry/myapp:latest

# Verify with annotations
cosign verify --key cosign.pub \
  -a build_id=123 \
  myregistry/myapp:latest
```

## Troubleshooting

### Common Issues

1. **"cosign: command not found"**
   - The container image includes cosign binary
   - Use the latest plugin image: `plugins/docker:latest`

2. **"keyless signing not supported"**
   - This plugin only supports private key signing
   - Don't use `--oidc` or `--identity-token` in `cosign_params`

3. **"encrypted private key requires password"**
   - Set `PLUGIN_COSIGN_PASSWORD` environment variable
   - Or use an unencrypted private key

4. **Registry authentication issues**
   - Cosign uses the same Docker registry credentials
   - Ensure Docker login is working first