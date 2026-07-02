---
name: session-reflection
description: "Conduct a two-question session-end reflection to catch blind spots and under-investigated areas before concluding. The agent surfaces its lowest-confidence work items and identifies what the user might be missing, then offers to investigate. Based on a Reddit-post technique combining an LLM-suggested confidence audit with Sam Altman's blind-spot question. Triggers: 'wrap up', 'we're done', 'conclude', 'session end', 'final review', 'before we go', 'sign off', 'that's all', 'anything else', 'finished', 'reflection', 'confidence check', 'blind spot', 'what are you missing', 'rate your confidence', 'review the session'."
---

# Session-End Reflection

Catch blind spots and under-investigated areas before concluding a session by asking two questions:

1. **Confidence audit:** What am I least confident about right now?
2. **Blind-spot check (Sam Altman):** What's the biggest thing I'm missing about this situation? What don't I realize?

~1 in 4 sessions, one of the answers reveals a critical gap that would silently invalidate work. This skill catches those gaps at the cheapest possible moment.

## Prerequisites

- A session that appears to be concluding (user signals completion, asks for summary, or starts wrap-up language)
- The `.agents/RULES.md` rule "Always conduct session-end reflection" should already be active (it references this skill)
- For persisting uncovered findings: the `context-file` skill, and optionally the `adr-capture` skill if a binding decision emerges

## Quick Start

```bash
# No commands needed — this is a behavioural skill.
# The rule in .agents/RULES.md triggers the reflection automatically.
# Read this skill for detailed guidance on execution.
```

## Workflow

### 1. Detect session-end signals

Look for cues like:
- User says "we're done", "thanks", "that's all", "wrap it up", "sign off"
- User asks for a summary or next steps
- User hasn't sent a message for a while (potential silence = conclusion)
- All identified tasks are marked complete

### 2. Initiate the reflection

Use a natural opening, for example:

> "Before we wrap up, I'd like to do a quick reflection. Two questions that often catch blind spots:"

Do NOT ask both questions at once. Ask sequentially, wait for the user's response to each.

### 3. Question 1: Confidence audit

Ask: **"What am I least confident about right now?"**

Generate 3–7 specific items. For each item:

| Item | Confidence | Why |
|------|-----------|-----|
| Assumed X without checking | Low | Only searched one location |
| Did not test edge case Y | Medium | Skipped due to time |
| Dependency version may be stale | Low | Did not verify registry |

Present this as a structured list. Be precise about what was under-investigated and why.

### 4. Question 2: Blind-spot check

Ask: **"What's the biggest thing I'm missing about this situation? What don't I realize?"**

This targets the user's blind spots rather than the agent's. Identify:
- Assumptions the user may have stated that were not verified
- Alternatives that were not explored
- Signals in the conversation that were noted but not followed up
- Constraints or requirements that were stated once but may have changed

### 5. Follow up

After the user responds to both questions:

- If the user flags an item for investigation, do deep root-cause investigation before concluding — search code, check docs, trace dependencies, test assumptions
- If an item reveals a genuine issue, offer to fix it before signing off
- If an item is a false alarm, explain why and move on
- If a finding warrants preservation, create a `.context/findings/` entry using the `context-file` skill

### 6. Conclude

Only after the investigation loop is resolved should the session end. If new work was spawned, note it clearly.

## When NOT to Use

- During a brief query that is clearly complete (e.g., "what's the capital of France?") — the reflection overhead is not justified
- When the user has explicitly said "don't do the reflection this time" or similar
- In automated/CI contexts — this is a human-interactive skill only
- In the middle of active work — only at session-end boundaries

## Reference: Question Design

| Aspect | Confidence Audit | Blind-Spot Check |
|--------|-----------------|------------------|
| Origin | LLM-suggested | Sam Altman |
| Focus | Agent's own work quality | Shared understanding |
| Scope | Under-investigated items | Assumptions & alternatives |
| Depth | 3–7 specific items | 1–3 broad patterns |
| Risk caught | Silent failures in execution | Conceptual blind spots |
| Frequency of critical finds | ~1 in 4 sessions | Not well documented but complementary |

## Anti-Patterns

### NEVER: Skip the reflection because the session feels complete

**WHY:** The whole point is that the agent cannot reliably assess its own completeness. The ~1-in-4 statistic means the agent is a poor judge of its own blind spots.

**BAD:** "Everything looks good, no need for reflection."
**GOOD:** Always run the reflection regardless of how confident the session seems.

### NEVER: Give vague confidence items

**WHY:** "I'm not confident about the overall approach" is not actionable. The user cannot decide whether to investigate without specifics.

**BAD:** "I'm least confident about performance."
**GOOD:** "I'm least confident about the query performance in `getUserOrders()` — I assumed the index exists but didn't verify against the production schema."

### NEVER: Ask both questions in one breath

**WHY:** The user needs space to think and respond to each question independently. Combining them reduces thoughtfulness of answers.

**BAD:** "What am I least confident about and what am I missing?"
**GOOD:** Ask, wait for full answer and discussion, then ask the second.

### NEVER: Deflect or make excuses for low-confidence items

**WHY:** The reflection is a safe space for surfacing uncertainty. Making excuses ("I would have checked if I'd had time") undermines the psychological safety needed for honest answers.

**BAD:** "I'm not confident about X, but that's because you didn't ask for it."
**GOOD:** "I'm not confident about X — I didn't verify it. Want me to check now?"

## Mindset

- The agent is a poor judge of its own blind spots — this is why the reflection is structured and mandatory
- Precision over quantity for confidence items. Better 3 specific items than 7 vague ones
- The blind-spot check is the harder question — it requires synthesizing across the entire session
- If a reflection reveals a critical issue, the session was not actually over — treat it as continuation, not wrap-up
- Persist important findings to `.context/findings/` so future sessions benefit from the discovery
- This technique works because the questions are complementary: internal audit (confidence) + external audit (blind spot)

## Troubleshooting

| Situation | Response |
|-----------|----------|
| User says "no need" to reflection | Accept gracefully. Do not insist. |
| User asks you to skip on a future session | Honour the preference. Consider noting it in `.context/` if project-level. |
| Reflection reveals a huge issue | Do not panic. Investigate calmly, present findings, offer remediation options. This is a win — you caught it before sign-off. |
| User has no response to either question | Accept that the reflection ran. The act of surfacing items is valuable even without follow-up. |

## Integration with Other Skills

| Skill | How it connects |
|-------|----------------|
| `context-file` | Persist reflection findings as `.context/findings/` entries when they reveal actionable gaps |
| `adr-capture` | If a reflection reveals a decision-level blind spot, capture it as an ADR |
| `rules-management` | If reflection reveals a pattern worth codifying as a behavioural rule, create one |

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Technique origins and rationale | [Session-End Reflection Reference](references/session-reflection-reference.md) | Understanding why these two questions work and how they complement each other |
