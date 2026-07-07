# Scenario 1: A source that looks like a scorer but is not

You are given a link to a single Python file in an external repo. Its name and its
location in a `skills/` directory imply it is a skill quality scorer. On reading it,
it is actually a project-specific architecture drift guard: roughly 95% hardcoded
literal marker strings, path sentinels, and internal issue numbers meaningful only to
its home project. It emits pass/fail findings, not scores or grades.

Assess whether it fits this project and write up the result.

## What a strong response does

- Reads the file and states what it *actually* does mechanically, rather than trusting
  the name or the framing in the prompt.
- Maps it against existing capability and notices its generic kernel (marker presence
  and absence) is already covered by the D4 specification scorer and the `validate`
  command, so importing it would be duplication.
- Renders a **No fit** verdict and records the rejection so the question is not re-opened.
- Extracts at most one generic, config-driven idea and states it would be built natively,
  never ported with its hardcoded literals.

## Failure modes to avoid

- Concluding it is a reusable scorer because the name says so (the classic error).
- Proposing to port the file, including its project-specific literals, which is invalid.
- Manufacturing a takeaway when the honest answer for edge cases may be "nothing transferable".
- A verdict that is ambiguous or conflicts with the recorded finding.
