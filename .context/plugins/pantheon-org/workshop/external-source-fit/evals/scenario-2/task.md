# Scenario 2: A source that fills a real gap

You are given a small, generic library that measures something this project does not:
whether a skill's trigger description reliably fires on the intended prompts and does
not fire on unrelated ones. It is language-agnostic, config-driven, and carries no
project-specific literals. No existing dimension (D1-D9), the `validate` command, or the
duplication engine measures trigger reliability.

Assess whether it fits this project and write up the result.

## What a strong response does

- Characterises the source mechanically and confirms it is generic, not hardcoded.
- Checks overlap and confirms no existing capability already covers trigger reliability.
- Renders a **Good fit** verdict because it fills a real gap with proportionate effort.
- Decides the vehicle: a deterministic measurement belongs in the Go CLI, not a helper skill.
- Recommends drafting a plan to build it natively, and grades the finding accordingly.

## Failure modes to avoid

- Rejecting a genuine gap because the source is in the wrong language (a cost, not a blocker).
- Conflating novel with needed in the wrong direction and dismissing a real gap.
- Failing to state the adoption vehicle, leaving the recommendation ambiguous.
- A verdict that conflicts with the evidence or is not one of the three bands.
