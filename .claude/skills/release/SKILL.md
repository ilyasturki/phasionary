---
name: release
description: Bump version in flake.nix, commit, and tag the release
argument-hint: "[major|minor|patch]"
disable-model-invocation: true
allowed-tools: Read, Edit, Bash(git add *), Bash(git commit *), Bash(git tag *)
---

# Release Skill

Bump the project version, commit the change, and create a git tag.

## Instructions

1. **Parse bump type**: Use `$ARGUMENTS` as the bump type. If empty or omitted, default to `patch`. Valid values: `major`, `minor`, `patch`. If the argument is not one of these, stop and tell the user.

2. **Read current version**: Read `flake.nix` and extract the version from the line matching `version = "X.Y.Z"` (line 13).

3. **Compute new version**: Parse the version as semver (MAJOR.MINOR.PATCH) and bump:
   - `patch`: increment PATCH (e.g. 0.5.1 -> 0.5.2)
   - `minor`: increment MINOR, reset PATCH to 0 (e.g. 0.5.1 -> 0.6.0)
   - `major`: increment MAJOR, reset MINOR and PATCH to 0 (e.g. 0.5.1 -> 1.0.0)

4. **Update flake.nix**: Use the Edit tool to replace the old version string with the new one in `flake.nix`.

5. **Commit**: Run `git add flake.nix` then `git commit -m "chore: bump version to v{new_version}"`. Use a simple `-m "message"` flag — do NOT use heredocs or `$()` command substitution.

6. **Tag**: Run `git tag v{new_version}` to tag the new commit.

7. **Report**: Print the version change (e.g. `v0.5.1 -> v0.5.2`) and confirm the tag was created.
