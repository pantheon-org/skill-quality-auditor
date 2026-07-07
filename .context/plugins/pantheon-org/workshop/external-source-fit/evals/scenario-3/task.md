# Scenario 3: A source with one good idea inside an ill-fitting whole

You are given a large eval framework whose overall design does not match this project:
different runtime, different agent model, an execution harness this project has no use
for. Buried inside it is one measurement technique, a counterfactual with-skill versus
without-skill comparison, that is genuinely worth adopting.

Assess whether it fits this project and write up the result.

## What a strong response does

- Characterises the framework mechanically and separates the whole from the one idea.
- Checks overlap and confirms the whole is an ill fit while the single technique is not
  already covered.
- Renders a **Partial fit** verdict: record the transferable idea, build only the kernel
  if and when a concrete need appears.
- Never proposes importing the whole framework or porting its project-specific glue.

## Failure modes to avoid

- Rendering Good fit for the whole framework when only one idea transfers.
- Rendering No fit and discarding the salvageable idea entirely.
- Importing the framework wholesale, an invalid outcome that creates conflict with the
  existing architecture.
- A verdict that is not exactly one of the three bands.
