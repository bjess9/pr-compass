# Troubleshooting

## Authentication Issues

### No GitHub token found

```
Error: auth_token_missing: No GitHub token found
```

**Solutions:**

1. Set environment variable: `export GITHUB_TOKEN="ghp_xxxxx"`
2. Use GitHub CLI: `gh auth login`
3. Verify token format (starts with `ghp_`, `gho_`, `ghu_`, `ghs_`)

### Invalid token format

```
Error: auth_token_invalid: GitHub token format is invalid
```

**Check:**

- Token starts with valid prefix (`ghp_`, `gho_`, etc.)
- Token length > 20 characters
- No extra whitespace/newlines

### Permission denied

```
Error: auth_permission_denied
```

**Fix:**

- Add `repo` and `read:org` scopes to token
- Check organization access permissions
- Verify team membership for team mode

## Configuration Issues

### Config not found

```
No configuration found. Create ~/.prpilot_config.yaml
```

**Solution:**

```bash
cp example_config.yaml ~/.prpilot_config.yaml
# Edit with your settings
```

### Invalid mode

```
Error: config_mode_invalid: Invalid mode 'xyz'
```

**Valid modes:** `repos`, `organization`, `teams`, `search`, `topics`

### Missing required fields

```yaml
# repos mode requires:
repos: ["owner/repo"]

# organization mode requires:
organization: "company"

# teams mode requires:
organization: "company"
teams: ["team-name"]

# search mode requires:
search_query: "org:company is:pr is:open"

# topics mode requires:
topic_org: "company"
topics: ["backend"]
```

## API Issues

### Rate limit exceeded

```
Error: github_rate_limit
```

**Wait for reset or:**

- Use authenticated requests (higher limits)
- Reduce concurrent repository fetching

### Repository not found

```
Error: github_not_found: Repository not found
```

**Check:**

- Repository name spelling
- Access permissions to private repos
- Organization membership

### Network errors

```
Error: github_network_error
```

**Verify:**

- Internet connectivity
- GitHub API status (status.github.com)
- Corporate firewall/proxy settings

## Display Issues

### No PRs shown

**Possible causes:**

1. **All filtered out** - Check filter settings
2. **No open PRs** - Verify repositories have PRs
3. **Permission issues** - Check repository access

### Empty table

```yaml
# Adjust filtering:
include_drafts: true
exclude_bots: false
exclude_authors: []
exclude_titles: []
```

### Slow loading

- **Too many repos** - Consider filtering
- **Network latency** - Use topics/teams mode
- **Large organization** - Filter by teams

## Docker Issues

### Permission denied (Docker)

```bash
# Mount GitHub CLI config:
docker run -v ~/.config/gh:/root/.config/gh:ro pr-compass

# Or use environment token:
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN pr-compass
```

### Config not found (Docker)

```bash
# Mount config directory:
docker run -v $(pwd)/config:/root/.config pr-compass
```

## Debug Mode

### Enable verbose logging

```bash
# Set log level (when implemented):
export LOG_LEVEL=debug
./pr-compass
```

### Test configuration

```bash
# Verify config syntax:
./pr-compass --version  # Should work if binary is OK
```

### Test authentication

```bash
# Test GitHub CLI:
gh auth status

# Test token manually:
curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/user
```

## Performance Issues

### High memory usage

- **Cause:** Many repositories/PRs
- **Fix:** Use filtering, topics mode

### Slow startup

- **Cause:** Large organization
- **Fix:** Switch to `teams` or `topics` mode

### UI freezing

- **Cause:** Network timeout
- **Fix:** Check connectivity, restart application

## Common Fixes

### Reset configuration

```bash
rm ~/.prpilot_config.yaml
cp example_config.yaml ~/.prpilot_config.yaml
```

### Clear authentication

```bash
unset GITHUB_TOKEN
gh auth logout
gh auth login
```

### Rebuild binary

```bash
make clean
make build
```

### Test with minimal config

```yaml
mode: 'repos'
repos: ['octocat/Hello-World'] # Public repo
exclude_bots: true
```

## Getting Help

1. Check configuration syntax
2. Verify authentication
3. Test with public repository
4. Enable debug logging
5. File issue with error message + config (sanitized)

**Always include:**

- OS and Go version
- PR Compass version (`./pr-compass --version`)
- Error message (full)
- Configuration (remove sensitive data)
