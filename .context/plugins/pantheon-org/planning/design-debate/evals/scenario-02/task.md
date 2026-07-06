# Scenario 02: Three-Sided Debate with Real Implementation Surface

## User Prompt

"I want to migrate our CLI's config file format from YAML to TOML. Give me a proper
collegial review of this — I want real pushback, not a rubber stamp."

## Expected Behavior

1. Restate the decision precisely: "Should the CLI's config file format be migrated from
   YAML to TOML?"
2. Investigate the actual repo: how many files/scripts currently parse the YAML config,
   whether any downstream tooling or user-facing docs assume YAML, and what a TOML
   migration would concretely touch.
3. Recognize this has real implementation surface (multiple consumers, migration cost)
   and add a third Migration/Risk role alongside Advocate and Skeptic — not just two
   roles.
4. Assign explicit opposing stances to Advocate and Skeptic; assign the neutral
   "assume it happens, find what breaks" framing only to the Migration/Risk role — not
   to all three.
5. Spawn all three in parallel.
6. Synthesize a verdict that specifically addresses what the Migration/Risk role found,
   not just the Advocate-vs-Skeptic tension.

## Success Criteria

- Agent recognizes the decision has real implementation surface and adds a third
  Migration/Risk role, per the skill's guidance to do so when a change touches multiple
  files/consumers.
- Advocate and Skeptic get genuinely opposing framings; only Migration/Risk gets a
  neutral "what would break" framing — the user's request for "pushback, not a rubber
  stamp" is satisfied by opposing roles, not by making every role adversarial toward the
  idea.
- All three reviewers are spawned in parallel.
- The synthesis explicitly incorporates a concrete finding from the Migration/Risk
  review, not just an Advocate/Skeptic summary.
- Agent does NOT reach for `plan-review` (no plan exists yet — this is a pre-plan idea).

## Failure Conditions

- Agent runs only 2 roles for a change with clear multi-file implementation surface,
  missing the Migration/Risk check.
- Agent gives all 3 reviewers the same "critique this" framing instead of distinct
  Advocate/Skeptic/Migration-Risk stances.
- Agent invokes `plan-review` or asks for a plan file, mistaking this for a
  post-plan review.
- The verdict ignores what the Migration/Risk reviewer specifically found.
