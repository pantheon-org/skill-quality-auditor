# Git Hooks Setup

Configure git hooks to run `skill-auditor` checks automatically. Below are examples using different hook managers.

## hk

```bash
# Install hk
curl -fsSL https://hk.jdx.dev/install.sh | sh

# Activate hooks in the repo
hk install
```

Example `hk.pkl`:

```pkl
amends "package://github.com/jdx/hk/releases/download/v1.48.0/hk@1.48.0#/Config.pkl"

local prePush = new Mapping<String, Step> {
  ["validate-artifacts"] {
    check = "skill-auditor validate artifacts"
  }
  ["check-duplication"] {
    check = "skill-auditor duplication <path>"
  }
  ["batch-audit"] {
    check = "skill-auditor batch <path> --fail-below B"
  }
}

hooks {
  ["pre-push"] { steps = prePush }
}
```

## pre-commit Framework

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: skill-validate
        name: Validate Artifacts
        entry: skill-auditor validate artifacts
        language: system
        pass_filenames: false
      - id: skill-duplication
        name: Check Duplication
        entry: skill-auditor duplication <path>
        language: system
        pass_filenames: false
      - id: skill-batch
        name: Batch Audit
        entry: skill-auditor batch <path> --fail-below B
        language: system
        pass_filenames: false
```

## Lefthook

```yaml
# lefthook.yml
pre-push:
  commands:
    skill-validate:
      run: skill-auditor validate artifacts
    skill-duplication:
      run: skill-auditor duplication <path>
    skill-batch:
      run: skill-auditor batch <path> --fail-below B
```

## Husky

```bash
# Install husky
npx husky init

# Add a pre-push hook
cat > .husky/pre-push << 'EOF'
#!/usr/bin/env sh
skill-auditor validate artifacts
skill-auditor duplication <path>
skill-auditor batch <path> --fail-below B
EOF
chmod +x .husky/pre-push
```

## Skipping Hooks

```bash
git commit --no-verify
git push --no-verify
```
