# Recommended Sub-Agent Models for Session Reflection

> **Caveat:** Prices below are sourced from [models.dev](https://models.dev/models/) (an open-source model database) and reflect the cheapest provider listed. What you actually pay depends on your infrastructure, provider contracts, and access tier. Prices change frequently. Always verify against your actual invoice.

For the sub-agent spawn pattern, the reflection task (reviewing a session summary, identifying blind spots, listing confidence gaps) does **not** need frontier reasoning. It benefits from:

| Requirement | Why |
|-------------|-----|
| Strong instruction following | Must produce structured output matching the prompt template |
| Large context window (≥256K) | Session summaries can be long with file paths, code snippets, command outputs |
| Tool-call support | Needed for the sub-agent to investigate flagged items (optional but valuable) |
| Reasoning capability | Must synthesize across the session, not just parrot the summary |
| Low cost per token | Reflection runs every session — should not materially increase bill |

## Recommended models (current as of mid-2026)

| Model | Est. Input/M tok | Est. Output/M tok | Context | Reasoning | Tool Call | Notes |
|-------|-----------------|------------------|---------|-----------|-----------|-------|
| DeepSeek V4 Flash | $0.00 | $0.00 | 1M | Yes | Yes | Open weights, widely available (31 providers) |
| Claude Sonnet 4.6 | $0.00 | $0.00 | 1M | Yes | Yes | Strong nuanced reasoning for blind-spot work |
| GPT-5.4 mini | $0.00 | $0.00 | 400K | Yes | Yes | Solid structured output |
| Qwen3.7 Plus | $0.00 | $0.00 | 1M | Yes | Yes | Huge context, good for long sessions |
| Step 3.7 Flash | $0.00 | $0.00 | 256K | Yes | Yes | Open weights, fast, 256K output capacity |
| Gemini 2.5 Flash-Lite | $0.07 | $0.28 | 1M | Yes | Yes | Largest context of any cheap model |
| GPT-5.4 nano | $0.18 | $1.10 | 400K | Yes | Yes | Ultra-cheap OpenAI option |
| Mistral Small 4 | $0.15 | $0.60 | 256K | Yes | Yes | Open weights, solid reasoning |

## What NOT to use

Avoid these for the reflection sub-agent:

- **No reasoning** models — the task requires synthesis and judgment; a pure feedforward model (e.g., older GPT-3.5-class, embedding models) cannot identify blind spots
- **No tool-call** models — limits ability to investigate flagged items autonomously
- **Nano-scale models** (sub-10B params without MoE) — typically too weak for reliable metacognitive reasoning, though this varies by architecture

## Provider caveat

The "$0.00" pricing on many models typically means a specific provider offers a free tier or promotional pricing. For example:
- DeepSeek's own API may charge while a third-party provider offers free credits
- Microsoft Azure might route OpenCode sub-agents differently from direct Anthropic access
- Your OpenCode/Claude Code environment may only have access to a subset of providers

**Always test your chosen model on 3-5 real session summaries before relying on it for the reflection.** A model that costs nothing but misses blind spots is more expensive than one that costs $0.15/M tok and catches them.
