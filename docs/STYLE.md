# Style

Conventions for everything under `docs/`, adapted from the
[source methodology](https://gist.github.com/lukewilson2002/cb48062397d8b51954034d94b8c19d6d) for
this fork.

## Writing

- Clear, direct Markdown. Write for a reader deciding whether to keep reading, not for
  completeness.
- Stable headings — other docs and code comments may link to a specific section; don't rename
  headings casually.
- Use relative Markdown links for related docs and source files, so they survive repo moves and
  work in both GitHub and local viewing.
- Focus on behavior, responsibilities, flows, invariants, and pitfalls — not a restatement of
  what the code already says line by line.
- Mark uncertainty explicitly (`(unconfirmed)`, "as far as we've traced, ...") rather than
  guessing and presenting it as fact. An honest gap is more useful than a confident wrong answer.
- Keep docs concise enough to read before making a change, not after.

## Frontmatter

Every doc under `systems/`, `flows/`, `architecture/`, and `architecture/decisions/` starts with
YAML frontmatter declaring an `audience` field — one or more persona slugs from
[`PERSONAS.md`](PERSONAS.md):

```yaml
---
audience: [contributor]
---
```

ADRs additionally require `status` and `date` (and `superseded_by` when superseded) — see
[`architecture/decisions/README.md`](architecture/decisions/README.md).

## Source maps

System and flow docs must include a source map: links to the files that matter most for that
system/flow (entry points, state definitions, handlers, tests). Don't enumerate every file
unless the system is genuinely small — prioritize the files a reader would otherwise have to
hunt for.

## Relationship to the Hugo site

If a doc's topic overlaps a page on the public Hugo site
(`site/content/en/docs/`), link to that page instead of restating its content — see
[`docs/README.md`](README.md#this-is-not-the-public-docs-site). Internal docs should focus on
the "how it works" / "why it's built this way" angle that the public site deliberately doesn't
cover.

## Code comments

Docs under `docs/` explain systems and flows; code comments explain *why a specific line of code
does something non-obvious* (ordering constraints, retry/lock/cache-invalidation reasons,
security-check rationale, edge cases being guarded against). Don't duplicate system-doc content
in a comment, and don't leave long architectural explanations inline in source files — link to
the doc instead.

## Keeping docs valid

A doc is only valid if it accurately describes the current code's intended behavior. When code
and docs disagree, either fix the doc, fix the code, or (if neither is possible on the spot) call
out the mismatch explicitly rather than leaving both to silently drift.
