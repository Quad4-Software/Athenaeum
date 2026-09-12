---
name: testing
description: >-
  How to write tests that actually catch bugs: spec-derived assertions,
  red-first discipline, boundary cases, mock discipline, property and
  mutation testing. Consult before writing or reviewing any test.
---

# Testing

A passing test and a useful test are different things. Agent-generated tests
start from the code, infer what it currently does, and assert that it does
that. Every test passes on the first run, coverage climbs, and the suite can
still miss most real bugs. In one documented run, 91 agent-written tests gave
100% line coverage and a 31% mutation score: two thirds of deliberately
injected bugs passed the suite.

The core rule: a test must be an independent statement of intent. Given this
input, the answer must be X, where X comes from the spec, the ticket, or the
contract; never from reading the code under test. If you cannot describe a
bug the test would catch, it is not a test.

## Fake-test anti-patterns

These produce green suites that verify nothing. Do not write them; flag them
in review.

### 1. The tautology test

The expected value is produced by the same logic being tested, or sits in the
mock setup and is echoed back.

```python
mock_repo.get.return_value = {"id": 7, "name": "Ada"}
result = svc.fetch_user(7)
assert result == {"id": 7, "name": "Ada"}   # asserts the mock works
```

This passes if `fetch_user` returns the repo row. It also passes if
`fetch_user` additionally deletes the database. **Assertion laundering** is
the subtler form: the test computes its expected value by calling the same
helper or transform the implementation uses. Both sides of the assertion run
identical logic, so the test can only fail if the language breaks.

### 2. Implementation-derived oracles

The test reads the function to decide what "expected" means. If the code has
a bug (returns `None` where it should raise), the test asserts `None`. The
bug is now documented, defended by CI, and the next person who fixes it sees
a red build and reverts. Derive expectations from the requirement, not the
implementation.

### 3. Happy-path-only inputs

Bugs live at boundaries: `<` vs `<=`, empty list, zero, the 32nd of the
month, the timezone-naive datetime, the oversized upload. Docstrings describe
the middle of the domain, so tests built from docstrings test the middle.
Every function with meaningful input ranges needs boundary cases.

### 4. Snapshots treated as truth

A snapshot written by something that never saw the correct output is a
photograph of current state. It detects change, not correctness. Snapshots
have one legitimate job: pinning behavior during refactoring. Label them for
that and do not count them as correctness tests.

### 5. Over-mocking

Mocking your own modules replaces every real seam (serialization, DB
constraints, clock behavior) with a stub that always cooperates. Measured
agent behavior: mocks appear in 36% of agent test commits vs 26% for humans.
A mock that returns exactly the shape the implementation expects restates the
code's assumptions; if the real dependency drifts, the test never notices.

### 6. Smoke tests counted as real tests

"Call it, assert the result is not null." This executes lines and checks
nothing. If it exists for a legitimate reason, name it what it is
(`TestXDoesNotPanic`) and do not count it toward coverage of behavior.

### 7. Coverage as the goal

Coverage measures which lines ran, not which lines were checked. A test that
executes a function and discards the result lights up every line. Optimizing
for coverage produces exactly the anti-patterns above. The metric to care
about is the mutation score: what fraction of deliberately broken code the
suite catches.

## What a real test looks like

### Spec-derived expected values

The expected output comes from the requirement: a literal you can justify
from the ticket, API contract, docstring contract, or domain invariant. If
you had to read the function body to write the expectation, stop; you are
writing a mirror.

### Red-first discipline

A test must be observed failing before the code that satisfies it exists.
When fixing a bug: write the test that fails on the pre-fix code, watch it
fail, then apply the fix. If you cannot make the test go red on the old code,
you have not understood the bug.

### Behavior at the boundary, not internals

Test what callers observe: inputs, outputs, errors, state changes visible
through the public API. Testing private helpers directly couples the suite
to the implementation and makes refactors expensive. If a private function
needs its own test, that is often a sign it wants to be a package of its own.

### Edge cases by default

For each function under test, consider: empty inputs, zero, one, maximum,
boundary values on both sides of a comparison, invalid input, error paths,
partial failure, and concurrency where the code has shared state.

### Mock only true external boundaries

Legitimate mock targets: the network, the clock, the filesystem, a payment
processor, an OIDC provider. Do not mock the code under test's collaborators
when a real or faked version is cheap. In this repo, SQLite runs in-process;
prefer a real in-memory database over a mocked store. Assert on outcomes,
not on "the mock was called with", unless the call itself is the contract
(e.g. an outbound webhook).

## Methods and patterns

### Table-driven tests (Go)

The repo standard is stdlib `testing` with table tests; no testify. See
`internal/auth/apikey_test.go` for the shape: a slice of cases with inputs
and `want` values, one loop, `t.Errorf` with inputs and got/want. Add a row
per edge case.

### Property-based tests

State an invariant; let the framework find the breaking input.

- Go: `testing/quick`, files named `*_property_test.go`, tests named
  `TestProperty*` (see `internal/auth/password_property_test.go`). Run with
  `task test:property`.
- Web: `fast-check` in Vitest (see `web/src/lib/utils/property.test.ts`).

Good invariants here: normalize-then-validate is idempotent, encode-then-
decode is identity, totals are never negative, ordering survives a round
trip, bounds always hold (`MinLength` stays within `[4, 128]`).

### Fuzz targets

Go fuzz tests live in `*_fuzz_test.go` files (e.g.
`internal/auth/apikey_fuzz_test.go`). Run a short pass with `task test:fuzz`
(default 10s per target), longer with `task test:fuzz:long`. Fuzz anything
that parses untrusted input: filenames, OPDS feeds, EPUB metadata, auth
tokens.

### Mutation testing

Mutation testing makes small changes to the source (flip `>` to `>=`, swap
`+` for `-`, drop a conditional) and reruns the suite. A mutant that still
passes "survived": no test cared about that line. This is the metric
coverage pretends to be.

- Go: `task test:mutation` (Gremlins)
- Web: `task test:mutation:web` (Stryker)

Run it on the module you changed, not the whole repo; it reruns the suite
per mutant. Workflow that works: run mutation testing, take each surviving
mutant, and write a test that fails when that mutant is applied. That is a
concrete pass/fail target instead of a vague quality goal.

### Contract and drift tests

`task test:contract` runs the OpenAPI/route/i18n/env drift checks. After
changing API routes or response shapes, run `task generate` then
`task test:contract`.

### Race and determinism

- `task test:race` for Go concurrency.
- Tests must be deterministic: inject clocks, seed randomness, use
  `t.TempDir()`, no real network, no fixed ports, no sleeps as
  synchronization.

## Commands

```sh
task test               # Go + Vitest
task test:go            # go test ./... (PKG=./internal/auth narrows)
task test:web           # vitest run
task test:race          # Go race detector
task test:property      # Go + web property suites
task test:fuzz          # short fuzz pass
task test:contract      # OpenAPI / route / i18n / env drift
task test:mutation      # Gremlins on Go code
task test:mutation:web  # Stryker on web code
task test:e2e           # Playwright (builds the binary first)
task test:coverage      # Go coverage report
```

## Review filter for any test (yours or generated)

Five-second checks per test:

1. If I delete the assertion, does the test still pass? It is a smoke test;
   name it that and stop counting it as coverage.
2. Does the expected value appear in the setup, or is it computed by the
   code under test? Tautology; rewrite with a spec-derived literal.
3. Would the test fail if I inverted one comparison or changed one constant
   in the function? If you cannot answer instantly, the behavior is
   untested.

Deeper rubric (from the VibeCheck study, arXiv 2609.05978): runnability,
assertion strength, edge-case coverage, isolation/determinism,
maintainability. Generated tests typically pass runnability and fail the
other four. Check the other four.

## What not to test

- Getters, setters, and pass-throughs with no logic.
- The framework or library itself (that a router routes, that bcrypt hashes).
- Exact log strings and error wording unless they are part of the contract.
- Private internals, when the same behavior is reachable through the public
  API.
