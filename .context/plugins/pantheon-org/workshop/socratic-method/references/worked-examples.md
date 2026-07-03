# Worked Examples

## Example 1: Architecture Decision

**User:** "I need to pick a database for our new service — should I use Postgres or DynamoDB?"

**Assistant (Phase 1):** "What kinds of queries does this service need to support? Read-heavy,
write-heavy, or a mix?"

**User:** "Mostly reads with complex joins across a few tables."

**Assistant (Phase 2):** "You mentioned 'new service' — are there existing services in your
org that already use one of these databases? Is consistency more important than latency here?"

**User:** "We're mostly a Postgres shop. Consistency matters — financial data."

**Assistant (Phase 3):** "If you went with DynamoDB, you'd need to redesign your join patterns
into single-table or multiple-query designs. Does your team have experience with that?"

**User:** "No, we're all SQL people."

**Assistant (Phase 4):** "What would happen if you used Postgres and hit a read bottleneck
six months from now?"

**User:** "We'd add read replicas. That's well-understood."

**Assistant (Phase 5):** "So the real constraints are: team expertise in SQL, financial data
requiring consistency, and manageable read-scale through replicas. Postgres seems like the
natural choice — should I proceed with that assumption?"

**User:** "Yes, let's go with Postgres."

## Example 2: Feature Request

**User:** "Add authentication to the app."

**Assistant (Phase 1):** "What kind of authentication — B2C users signing up, or enterprise
SSO, or both?"

**User:** "Both eventually, but start with B2C."

**Assistant (Phase 2):** "When you say 'B2C' — are you assuming email/password, or do you
want social login options too?"

**User:** "Social login would be nice but we have no experience with it."

**Assistant (Phase 3):** "If we add social login later, the data model needs to support
multiple identity providers from the start. That adds some up-front complexity."

**User:** "Let's keep it simple — email/password first."

**Assistant (Phase 4):** "What if a user forgets their password? How should we handle
password resets?"

**User:** "Email-based reset, standard flow."

**Assistant (Phase 5):** "So: start with email/password for B2C, design the user model to
support multiple providers, implement standard password reset. Shall I proceed?"

**User:** "Yes."
