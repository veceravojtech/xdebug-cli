# Change: Add Post-Apply Install Step

## Why
After applying a proposal that modifies CLI code, the installed binary in `~/.local/bin/xdebug-cli` becomes outdated. Developers must manually remember to run the install command. Adding a mandatory post-apply step ensures the environment always has the latest version.

## What Changes
- Update `.claude/commands/openspec/apply.md` to add step 5: Run `./install.sh` to update the installed binary
- Add verification step to confirm installation succeeded

## Impact
- Affected specs: None (tooling-only change)
- Affected code: `.claude/commands/openspec/apply.md`
