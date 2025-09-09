# Release Guide

This document explains how to create a new release for the Perplexity MCP Golang server.

## Release Process

The project uses an automated GitHub Actions workflow to handle releases. This ensures consistency and reduces manual errors.

### Prerequisites

1. Ensure your changes are merged into the main branch
2. Verify all tests pass
3. Review the commit history since the last release

### Creating a Release

1. Go to the **Actions** tab in the GitHub repository
2. Select the **Release** workflow from the left sidebar
3. Click **Run workflow** button
4. Fill in the version field using semantic versioning format (e.g., `v1.0.0`)
5. Click **Run workflow** to start the process

### What the Workflow Does

The automated release workflow performs the following steps in order:

1. **Validation**: Checks version format and ensures the tag doesn't already exist
2. **Git Tag**: Creates and pushes the version tag to the repository
3. **Draft Release**: Creates a draft GitHub release with auto-generated notes
4. **Docker Build**: Builds multi-architecture Docker images (amd64, arm64)
5. **Docker Push**: Pushes images to GitHub Container Registry with multiple tags:
   - `ghcr.io/owner/repo:v1.0.0` (version tag)
   - `ghcr.io/owner/repo:latest` (latest tag)
   - `ghcr.io/owner/repo:abc123` (commit SHA tag)
6. **Publish Release**: Publishes the GitHub release with updated Docker pull instructions

### Failure Recovery

If the workflow fails at any step:

- **Before Docker push**: Only the Git tag and draft release are created. You can delete the tag and draft release, fix the issue, and retry.
- **During Docker push**: The Git tag and draft release exist, but no Docker images are published. Fix the issue and retry.
- **After Docker push**: Everything is successfully released.

This order minimizes the risk of partial releases and makes cleanup easier if something goes wrong.

### Version Format

Use semantic versioning (SemVer) format:
- `v1.0.0` for major releases
- `v1.1.0` for minor releases  
- `v1.0.1` for patch releases

### Using Released Images

After a successful release, users can pull the Docker image using:

```bash
# Pull specific version
docker pull ghcr.io/owner/repo:v1.0.0

# Pull latest version
docker pull ghcr.io/owner/repo:latest

# Pull specific commit
docker pull ghcr.io/owner/repo:abc123
```

### Troubleshooting

Common issues and solutions:

- **"Tag already exists" error**: The version tag has already been created. Use a different version number.
- **Docker build fails**: Check the Dockerfile and ensure it builds locally first.
- **Permission denied**: Ensure the repository has proper permissions for GitHub Container Registry.

### Manual Cleanup

If you need to clean up a failed release:

1. Delete the Git tag: `git tag -d v1.0.0 && git push origin :refs/tags/v1.0.0`
2. Delete the draft release from the GitHub web interface
3. Remove Docker images from the Container Registry if they were pushed