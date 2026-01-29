# Claude Code Skills

This directory contains specialized skills for Claude Code.

## Available Skills

### golangci-lint-best-practices

**Purpose**: Generate and fix Go code following golangci-lint best practices

**Usage**:
```
@golangci-lint-best-practices Generate a function to process events
```

**Features**:
- Reads `.golangci.yml` configuration
- Generates linter-compliant code
- Fixes linter violations
- Explains best practices
- Provides before/after examples

**Documentation**: `../docs/GOLANGCI_SKILL_GUIDE.md`

---

## How Skills Work

Claude Code automatically discovers and uses skills when:
1. You explicitly invoke them: `@skill-name <request>`
2. Your request matches the skill's domain
3. The skill is relevant to the current task

---

## Creating New Skills

To create a new skill:

1. Create a markdown file in `.claude/skills/`
2. Define the skill's purpose and capabilities
3. Provide instructions and examples
4. Document when the skill should be used

Example structure:
```markdown
# My Skill Name

You are a specialist in [domain].

## Capabilities
- What the skill can do

## Instructions
- How to use the skill

## Examples
- Concrete examples

## Notes
- Important considerations
```

---

## Best Practices

- **Specific**: Skills should have a clear, focused purpose
- **Documented**: Include examples and use cases
- **Context-aware**: Skills can read project files
- **Composable**: Skills can work together
- **Version controlled**: Commit skills to git for team sharing

---

## Skill Invocation

### Explicit
```
@golangci-lint-best-practices Fix the errors in this code
```

### Implicit (Claude decides)
```
Generate a new HTTP handler following our coding standards
```

### In Conversation
```
Can you use the golangci-lint skill to help me refactor this function?
```

---

## Sharing Skills

Skills in this directory are:
- ✅ Version controlled (commit to git)
- ✅ Shared with entire team
- ✅ Automatically discovered by Claude Code
- ✅ Updated when you pull latest code

---

## Current Project Skills

| Skill | Purpose | Status |
|-------|---------|--------|
| golangci-lint-best-practices | Generate linter-compliant Go code | ✅ Active |

---

For more information on Claude Code skills, see: https://claude.ai/docs/code/skills
