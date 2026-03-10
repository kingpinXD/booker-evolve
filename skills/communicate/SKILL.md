# Skill: Communicate

Formats and conventions for session logging, learnings, and issue responses.

## JOURNAL.md Format

Append one entry per session at the bottom of JOURNAL.md:

```markdown
## Day N -- HH:MM -- [descriptive title]

[2-4 sentences summarizing what was done, what succeeded, what failed, and any decisions made.]
```

- Use the current DAY_COUNT for N
- Use 24-hour UTC time for HH:MM
- Be honest about failures — they are more valuable than successes
- Keep it factual: what happened, not what you hoped would happen

## LEARNINGS.md Format

Append entries when you discover something generalizable:

```markdown
## Lesson: [one-line insight]

[Paragraph explaining the context: what happened, why it matters, and when this lesson applies. Include specific examples from the session.]
```

Only add a learning if it will be useful in future sessions. Do not log obvious things.

## ISSUE_RESPONSE.md Format

When a GitHub issue was addressed during the session, write:

```markdown
## Issue #N: [title]

**Status:** [investigating|fixed|wontfix|needs-info]
**Commit:** [short SHA of the commit that fixes it, e.g. abc1234]
**Summary:** [1-2 sentences explaining what was done]
**Changes:** [list of files modified, or "none" if just investigated]
```

- If Status is `fixed`, include the commit SHA. The orchestrator will close the issue automatically.
- Multiple issues can be listed in the same file.
- This file is ephemeral — it gets parsed by evolve.sh to post comments and close fixed issues.

## Style Rules

- Concise and technical
- No emojis
- State facts, not feelings
- Acknowledge uncertainty explicitly
