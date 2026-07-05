# Scenario 02: Adaptive Branching and Recap

## Setup

A guided interview is already underway about choosing a deployment model for a new service.
The agent had planned to ask about rollout strategy (blue-green vs. canary vs. rolling) as its
next question. The user's most recent answer was to the first question:

## User Prompt (answer to the first question)

"We're going serverless — no long-running instances at all."

## Expected Behavior

1. Recognize that "serverless" makes the planned rollout-strategy question (blue-green/canary/rolling, which assumes long-running instances) moot.
2. Do not ask the originally planned question unchanged — adapt to a question that fits a serverless context instead (e.g. cold-start tolerance, concurrency limits, or deployment frequency).
3. Continue asking only one question at a time.
4. Once enough answers are gathered (this scenario ends after 2-3 total questions), present a short bulleted recap of all answers so far.
5. Ask the user to confirm the recap is accurate before finalizing anything.

## Success Criteria

- The next question asked is visibly adapted to "serverless" rather than being the originally-implied rollout-strategy question.
- Only one question is asked per turn throughout.
- A recap is presented as a bulleted list reflecting the answers given.
- The agent explicitly asks for confirmation of the recap before producing a final answer or document.

## Failure Conditions

- The agent asks the blue-green/canary/rolling question anyway, ignoring that it no longer applies.
- The agent bundles the next question with others.
- The agent produces a final recommendation or document without a recap and confirmation step.
- The recap misstates or drops the "serverless" answer.
