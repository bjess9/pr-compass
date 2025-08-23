# Docker Usage

## Run (Recommended)

```bash
# With GitHub CLI auth (most secure)
docker run --rm -v ~/.config/gh:/root/.config/gh:ro ghcr.io/bjess9/pr-compass:latest

# With token (less secure)
docker run --rm -e GITHUB_TOKEN=ghp_xxx ghcr.io/bjess9/pr-compass:latest
```

## Custom Config

```bash
# Mount your config file
docker run --rm \
  -v ~/.prpilot_config.yaml:/root/.prpilot_config.yaml:ro \
  -v ~/.config/gh:/root/.config/gh:ro \
  ghcr.io/bjess9/pr-compass:latest
```

## Development

```bash
# Use docker-compose for development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Access dev container
docker-compose exec pr-compass sh
```

**Troubleshooting**: See [docs/troubleshooting.md](docs/troubleshooting.md#docker-issues)
