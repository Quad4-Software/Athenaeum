---
name: code-quality
description: >-
  How agents should write code here: match conventions, smallest correct
  change, no speculative abstraction, no dead code, readable over clever.
  Consult before writing or refactoring any code.
---

# Code quality for agents

Generated code measurably underperforms human code on structure. Studies of
LLM output find roughly 60% more code smells than reference solutions, with
implementation smells dominating and the gap widening as tasks get more
complex (arXiv 2510.03029). A 2026 analysis of agent-driven development
identified a "machine signature" of debt: complex logic collapsed into long
procedural methods, god classes at system scale, hub-like coupling, and a
near-linear link between generated volume and architectural decay (arXiv
2605.02741). Correctness and quality are decoupled: code that passes tests is
just as likely to be structurally flawed as code that fails.

This skill exists to counter that signature. Every rule below targets a
documented failure mode of generated code.

## 1. Match the codebase before writing anything

- Read neighboring files in the same package or directory first. Copy their
  patterns for errors, logging, naming, file layout, and imports.
- Reuse existing helpers, constants, and types. Search for what you are about
  to write; this repo shares `internal/config`, `internal/models`, and utility
  packages for a reason.
- Never assume a library is available. Check `go.mod`, `package.json`, or the
  imports of sibling files before using one.
- If the codebase does it "wrong" consistently, match it anyway or fix it
  everywhere in a separate change. Do not mix two conventions in one diff.

## 2. Smallest correct change

- Implement what was asked. Do not refactor adjacent code, rename things, or
  reformat lines you did not need to touch.
- A diff that changes 20 files for a 2-file fix is a smell, not diligence.
- Leave working code alone. Resist the urge to "improve" code you happen to
  be reading.

## 3. No speculative abstraction

- No interfaces with one implementation "in case we swap it later".
- No options, flags, or parameters for hypothetical requirements.
- No generic utilities before the second concrete use case exists. The first
  copy can stay a copy; the third copy earns a helper.
- No wrapper functions that add a name but no behavior.

## 4. Duplication and reuse

- Before writing a helper, grep for it. Similar code already exists in most
  mature repos.
- Extract shared helpers into the right shared location, not the file you
  happen to be in. One-off copies of the same logic in N files is how
  generated codebases rot.
- Constants over magic strings and numbers: shared paths, query keys, route
  fragments, timeouts, status strings.

## 5. Size and shape

- No god files. When a file grows past comfortable review size, split by
  concern into the same package or a focused new one.
- Functions do one thing at one level of abstraction. If you need the word
  "and" to describe what a function does, it is two functions.
- Early returns over nested conditionals. Keep nesting shallow; past about
  three levels, restructure.
- The research finding to internalize: agents push complexity into a single
  long method when reasoning gets hard. When you feel a function growing past
  a screen, that is the signal to extract, not to continue.

## 6. Naming

- Names say what the thing holds or does: `failedLogins`, not `data`;
  `parseScrobble`, not `process`.
- Booleans read as predicates: `isAdmin`, `hasAccess`, `canWrite`.
- Match the codebase's vocabulary. If the domain calls it a "shelf", do not
  introduce "collection" for the same concept.
- No single-letter names outside tight loops, no abbreviations the codebase
  does not already use.

## 7. Comments

- Comments explain why, not what. The code already says what.
- No comments that narrate the diff or the task: "added error handling",
  "fixed bug", "new function". Delete these before finishing.
- No backtick code spans inside comments.
- Update or delete comments that your change makes stale. A wrong comment is
  worse than none.

## 8. Error handling

- Handle errors at the layer that can do something about them. Do not log
  and swallow; return or propagate.
- Match the package's existing error style: sentinel errors, wrapped errors,
  status codes. Do not invent a new scheme per file.
- Do not defensively check states that cannot occur; assert or document the
  invariant. Validation belongs at trust boundaries: user input, network,
  disk, database.
- Empty catch/except blocks are never acceptable. Neither is returning a
  default that silently hides a failure.

## 9. Dead code

- Delete unused functions, exports, parameters, and branches. Do not comment
  them out.
- This repo has tools for this: `task deadcode` (Go) and `task knip` (web).
  Run them when a refactor might have orphaned code.
- Remove unreachable branches added "for safety". If the type system or a
  prior check makes a case impossible, the branch is noise.

## 10. Volume is not quality

- The strongest predictor of architectural decay in generated code is lines
  written. More code is more debt, not more value.
- Deleting code is a valid outcome. A fix that removes a special case beats
  one that adds a flag.
- If a task can be done by deleting or reusing, prefer that over writing new
  code.

## 11. Security and correctness basics

- Never log secrets, tokens, or credentials. Never commit them.
- Parameterize queries; no string-built SQL with user input.
- Validate and bound user-controlled input at the boundary (sizes, counts,
  paths). Path traversal, injection, and unbounded allocation are the
  classic generated-code slips.
- Crypto and auth code: use the existing packages (`internal/auth`,
  `internal/oidc`); never hand-roll.

## Self-review before finishing

After writing code, re-read your own diff and check:

1. Could a reviewer tell what changed and why from the diff alone?
2. Did I touch anything the task did not require? Revert it.
3. Is there a helper or constant that already does this? Use it.
4. Any function that needs scrolling to understand? Split it.
5. Any abstraction with exactly one caller and no second use case? Inline or
   drop it.
6. Any dead code, commented-out blocks, or stale comments left behind?
7. Did I match the error and naming style of the package I touched?
8. Would the simplest version of this code still be correct? Prefer it.
9. Run the format and lint pass on touched files (`task fmt`, `task lint`
   or the narrower equivalents). Fix what you introduced.
