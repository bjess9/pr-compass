# Configuration

**Location**: `~/.prpilot_config.yaml`  
**Example**: See `example_config.yaml` for full syntax

## Mode Selection Guide

**`topics`** (Recommended): Catches external PRs to your repos. Best for team leads.

**`search`**: Most flexible. Use GitHub search syntax. Good for complex queries.

**`organization`**: Simple but can be overwhelming for large orgs.

**`teams`**: Requires team membership. May miss cross-team PRs.

**`repos`**: Manual maintenance. Good for small, fixed repo sets.

## Non-obvious Behaviors

**Bot filtering**: Hardcoded list in code. `exclude_bots: false` to disable.

**Title patterns**: Partial matching. `"chore:"` matches `"chore: update deps"`.

**Topics mode**: Finds repos by topics, then fetches ALL their PRs. Not topic-filtered PRs.

**Search queries**: Use GitHub search syntax. Rate limits apply differently.

## Performance Tips

**Large orgs**: Use `topics` or `teams` mode, not `organization`.

**Many repos**: Consider filtering with `exclude_titles` to reduce noise.

**Rate limits**: Authenticated requests have higher limits. Set `GITHUB_TOKEN`.
