# 🐳 Docker Support for PR Pilot

PR Pilot supports running in Docker containers for easy deployment and development. This guide covers all Docker-related usage scenarios.

## 🚀 Quick Start

### Pull from Docker Hub

```bash
# Pull the latest version
docker pull bjess9/pr-pilot:latest

# Run with GitHub token
docker run --rm -e GITHUB_TOKEN=your_token_here bjess9/pr-pilot:latest
```

### Build Locally

```bash
# Build the image
docker build -t pr-pilot .

# Run the container
docker run --rm pr-pilot --version
```

## 🔐 Authentication

PR Pilot supports multiple authentication methods in Docker:

### Environment Variable

```bash
docker run --rm \
  -e GITHUB_TOKEN=ghp_your_token_here \
  bjess9/pr-pilot:latest
```

### GitHub CLI (Recommended)

Mount your GitHub CLI configuration:

```bash
docker run --rm \
  -v ~/.config/gh:/root/.config/gh:ro \
  bjess9/pr-pilot:latest
```

### Configuration File

Mount a configuration directory:

```bash
docker run --rm \
  -v ./config:/root/.config \
  -e GITHUB_TOKEN=ghp_your_token_here \
  bjess9/pr-pilot:latest
```

## 🛠️ Development with Docker

### Using Docker Compose

For development, use the provided docker-compose setup:

```bash
# Development environment
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Access the development container
docker-compose exec pr-pilot sh

# Run tests inside container
docker-compose exec pr-pilot go test ./...

# Build the application
docker-compose exec pr-pilot go build -o pr-pilot ./cmd/pr-pilot
```

### Development Workflow

1. **Start development environment:**

   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
   ```

2. **Attach to development container:**

   ```bash
   docker-compose exec pr-pilot sh
   ```

3. **Make code changes** (files are mounted as volumes)

4. **Test changes:**

   ```bash
   go run ./cmd/pr-pilot --version
   go test ./...
   ```

5. **Clean up:**
   ```bash
   docker-compose down
   ```

## 📋 Configuration Examples

### Basic Usage

```bash
# Check version
docker run --rm bjess9/pr-pilot:latest --version

# List PRs with environment token
docker run --rm \
  -e GITHUB_TOKEN=ghp_your_token_here \
  bjess9/pr-pilot:latest
```

### Advanced Configuration

```bash
# Mount config file and run interactively
docker run -it --rm \
  -v $(pwd)/.prpilot_config.yaml:/root/.prpilot_config.yaml:ro \
  -v ~/.config/gh:/root/.config/gh:ro \
  bjess9/pr-pilot:latest
```

### Production Deployment

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  pr-pilot:
    image: bjess9/pr-pilot:latest
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    volumes:
      - ./config:/root/.config
    restart: unless-stopped
    networks:
      - app-network

networks:
  app-network:
    external: true
```

## 🏗️ Multi-Architecture Support

The Docker images are built for multiple architectures:

- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon, ARM servers)

Docker will automatically pull the correct architecture for your platform.

```bash
# Explicitly specify architecture if needed
docker run --platform linux/amd64 --rm bjess9/pr-pilot:latest --version
docker run --platform linux/arm64 --rm bjess9/pr-pilot:latest --version
```

## 📊 Container Health Checks

The container includes health checks:

```bash
# Check container health
docker ps
# CONTAINER ID   IMAGE                    COMMAND         CREATED         STATUS                    PORTS     NAMES
# abc123def456   bjess9/pr-pilot:latest   "./pr-pilot"    2 minutes ago   Up 2 minutes (healthy)             pr-pilot

# Manually run health check
docker exec <container_id> ./pr-pilot --version
```

## 🔍 Troubleshooting

### Common Issues

#### 1. Authentication Errors

```bash
# Problem: No authentication found
# Solution: Set GITHUB_TOKEN or mount GitHub CLI config
docker run --rm \
  -e GITHUB_TOKEN=your_token_here \
  bjess9/pr-pilot:latest
```

#### 2. Permission Issues

```bash
# Problem: Permission denied accessing mounted files
# Solution: Check file ownership and permissions
ls -la ~/.config/gh
chmod -R 600 ~/.config/gh
```

#### 3. Network Connectivity

```bash
# Test network connectivity from container
docker run --rm bjess9/pr-pilot:latest sh -c "ping -c 1 api.github.com"
```

### Debugging

#### 1. Interactive Shell

```bash
# Get shell access to debug issues
docker run -it --rm --entrypoint sh bjess9/pr-pilot:latest

# Check installed packages
apk list --installed | grep github
```

#### 2. Verbose Logging

```bash
# Enable verbose output (when feature is added)
docker run --rm \
  -e GITHUB_TOKEN=your_token_here \
  -e LOG_LEVEL=debug \
  bjess9/pr-pilot:latest
```

#### 3. Container Logs

```bash
# View container logs
docker logs <container_name>

# Follow logs in real-time
docker logs -f <container_name>
```

## 🏷️ Image Tags

Available image tags:

- `latest` - Latest stable release from main branch
- `v1.0.0`, `v1.0`, `v1` - Specific version tags
- `main` - Latest development build from main branch
- `develop` - Latest development build from develop branch

```bash
# Use specific version
docker run --rm bjess9/pr-pilot:v1.0.0 --version

# Use development version
docker run --rm bjess9/pr-pilot:develop --version
```

## 🔒 Security Considerations

### Token Security

1. **Use environment variables for tokens:**

   ```bash
   # Good: Environment variable
   docker run --rm -e GITHUB_TOKEN=ghp_xxx pr-pilot:latest

   # Avoid: Token in command line (visible in process list)
   docker run --rm pr-pilot:latest --token ghp_xxx
   ```

2. **Use Docker secrets in production:**

   ```yaml
   version: '3.8'
   services:
     pr-pilot:
       image: bjess9/pr-pilot:latest
       environment:
         - GITHUB_TOKEN_FILE=/run/secrets/github_token
       secrets:
         - github_token

   secrets:
     github_token:
       file: ./secrets/github_token.txt
   ```

3. **Mount GitHub CLI config as read-only:**
   ```bash
   -v ~/.config/gh:/root/.config/gh:ro
   ```

### Container Security

- Containers run as non-root user where possible
- Minimal Alpine Linux base image
- Regular security scanning with Trivy
- No sensitive data stored in image layers

## 🚢 CI/CD Integration

### GitHub Actions

```yaml
name: Run PR Pilot in Docker
on: [push, pull_request]

jobs:
  pr-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run PR Pilot
        run: |
          docker run --rm \
            -e GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }} \
            bjess9/pr-pilot:latest
```

### GitLab CI

```yaml
pr-pilot-check:
  image: bjess9/pr-pilot:latest
  script:
    - pr-pilot
  variables:
    GITHUB_TOKEN: $GITHUB_TOKEN
```

### Jenkins

```groovy
pipeline {
    agent {
        docker { image 'bjess9/pr-pilot:latest' }
    }
    environment {
        GITHUB_TOKEN = credentials('github-token')
    }
    stages {
        stage('Check PRs') {
            steps {
                sh './pr-pilot'
            }
        }
    }
}
```

## 📈 Monitoring and Observability

### Container Metrics

```bash
# Monitor resource usage
docker stats pr-pilot

# Check container resource limits
docker inspect pr-pilot | grep -A 20 "HostConfig"
```

### Health Monitoring

```bash
# Set up health checks in docker-compose
services:
  pr-pilot:
    image: bjess9/pr-pilot:latest
    healthcheck:
      test: ["CMD", "./pr-pilot", "--version"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

## 🤝 Contributing

When contributing Docker-related changes:

1. **Test both architectures:**

   ```bash
   docker buildx build --platform linux/amd64,linux/arm64 .
   ```

2. **Update documentation** for any new Docker features

3. **Test with different authentication methods**

4. **Verify security scanning passes**

For more information, see the main [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md).
