# Install Stack Explorer on a new machine

Steps to clone, build, and restore runtime state (DB + scans + reports) on a second machine.

## Prereqs

- Go 1.26+
- `sqlite3` CLI (for importing the DB dump)
- `git`
- Optional: `golangci-lint`, `scc`, Hadron CLI if you plan to run scans

## 1. Clone + build

```bash
git clone git@github.com:hollis-labs/stack-explorer.git
cd stack-explorer
make build          # -> ./stack-explorer
# or:
make install        # -> ~/go/bin/stack-explorer
```

## 2. Restore state from export tarball

Produced by `./scripts/export-state.sh` on the source machine.

```bash
# Extract the tarball somewhere (e.g. ~/Downloads)
tar -xzf ~/Downloads/stack-explorer-state-YYYYMMDD-HHMMSS.tar.gz
cd stack-explorer-state-YYYYMMDD-HHMMSS

cat MANIFEST.txt    # sanity check: DB size, repo count, source host

# Import DB
sqlite3 <path-to-stack-explorer>/data/stack-explorer.db < db.sql

# Restore scans + reports
cp -R scans/*   <path-to-stack-explorer>/data/scans/
cp -R reports/* <path-to-stack-explorer>/reports/

# Verify
cd <path-to-stack-explorer>
./stack-explorer db stats
```

`repos.yaml` is in the tarball for reference but you already have it from git — skip it unless you're overriding.

## 3. Path correction

The source machine had `/Users/chrispian/Projects-apps/` as the app root. On the new machine, replace it with your prefix everywhere it appears.

**Find all hardcoded paths:**

```bash
grep -rn "/Users/chrispian/Projects-apps" . \
  --include="*.yaml" --include="*.yml" --include="*.md" --include="*.sh"
```

**Bulk rewrite** (BSD sed — macOS):

```bash
NEW_PREFIX="$HOME/Projects-apps"   # adjust to taste
find . -type f \( -name '*.yaml' -o -name '*.md' -o -name '*.sh' \) \
  -not -path './.git/*' -not -path './tmp/*' \
  -exec sed -i '' "s|/Users/chrispian/Projects-apps|$NEW_PREFIX|g" {} +
```

**GNU sed (Linux):** drop the `''` after `-i`.

**Where these paths live:**

| File | What it points at |
|------|-------------------|
| `repos.yaml` (×11 `local_path:` lines) | Own-project clone paths for unreleased repos |
| `blueprints/se-repo-scan.yaml` | Default scan output directory |
| `skills/*.md` (×10) | Example commands referencing the repo root |
| `CLAUDE.md` | Hadron run example |
| `docs/stack-explorer-api-prompt.md` | Historical boot-spec reference |

**`local_path:` entries in `repos.yaml`** point at sibling repos that must also exist locally if you want to run scans against them. On a fresh machine without those siblings, those entries will fail scans but don't block anything else.

## 4. Verify

```bash
./stack-explorer db stats          # expect same counts as MANIFEST.txt
./stack-explorer repo list | head  # 113 repos
./stack-explorer score get conduit --lens chat-app  # sanity
```

## 5. Re-exporting later

When you want to sync state back the other way:

```bash
./scripts/export-state.sh   # writes ../stack-explorer-state-*.tar.gz
```

Transfer the tarball and repeat step 2 on the destination. The DB import is destructive — back up the target DB first if it has local-only changes.
