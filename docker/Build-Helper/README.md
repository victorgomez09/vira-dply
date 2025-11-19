# MikroCloud Builder

This is a comprehensive build helper container that includes all necessary tools for building applications in MikroCloud.

## Included Tools

- **Git** - Source code management
- **Docker CLI** - Container image building
- **Docker Compose** - Multi-container builds
- **Docker Buildx** - Advanced Docker builds
- **Nixpacks** - Automatic buildpack detection
- **Pack (Buildpacks)** - Cloud Native Buildpacks
- **MinIO Client** - Object storage operations

## Building Locally

```bash
# Build for local use
make build-helper

# Or manually
docker build -t mikrocloud-builder:latest ./docker/Build-Helper
```

## Pushing to GitHub Container Registry

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and tag
make build-helper

# Push to registry
make build-helper-push
```

## GitHub Actions Workflow

The `.github/workflows/build-helper.yml` workflow automatically:
- Builds the image on changes to `docker/Build-Helper/**`
- Builds for both `linux/amd64` and `linux/arm64`
- Pushes to `ghcr.io/fantasy-programming/mikrocloud-2/mikrocloud-builder:latest`
- Runs on push to `main` branch

## Usage in MikroCloud

The build service automatically uses this image for:
- Static site builds
- Dockerfile builds
- Docker Compose builds

The image is configured in `/pkg/containers/build/service.go`.

## Version Information

See the Dockerfile for specific versions of included tools:
- Alpine: 3.21
- Docker: 28.0.0
- Docker Compose: 2.38.2
- Docker Buildx: 0.25.0
- Pack: 0.38.2
- Nixpacks: 1.40.0
