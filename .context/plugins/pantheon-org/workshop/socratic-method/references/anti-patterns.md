# Anti-Patterns

## Question Dump

Asking all five phases' questions in a single turn. Violates the rule of three and
overwhelms the user. Diagnostic: your message contains four or more questions.

*Fix:* Ask the single most important question. Let the answer guide the next one.

## Leading Questions

Phrasing a question that telegraphs the expected answer. For example: "Don't you think
we should use X?" or "Wouldn't it be better to Y?".

*Fix:* Open questions should have no preferred answer. "What approaches have you considered?"
not "Have you considered X?".

## Solution Anchoring

Dropping a partial solution while questioning — "Maybe we could use X, but first tell me
about Y". Once the solution is named, the user anchors to it and stops exploring.

*Fix:* Zero implementation details until Phase 5 confirmation. If you must say something
helpful, say what you *don't* know yet.

## Persisting After Override

The user says "just do it" and the skill persists with "But have you considered...?".
This erodes trust and wastes time.

*Fix:* Acknowledge once ("Understood — proceeding on the original request"), note what
was skipped, and execute.

## Ceremonial Questioning

Applying the protocol to a straightforward task because the skill triggered. The user
asked for a specific thing and gets Socratic pushback.

*Fix:* The when-not-to-use list is a hard gate. If the task is concrete, skip the protocol.
