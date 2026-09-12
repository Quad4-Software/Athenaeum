# No-AI-Slop Writing Rules

## Purpose

Portable rules that keep agent-written prose free of the patterns that mark
machine-generated text. Applies to docs, README sections, UI copy, commit
messages, PR text, comments that are sentences, and chat replies.

The rules are about the writing itself: what makes prose specific, honest, and
free of tells. They are not tied to any wiki, CMS, or publishing pipeline.

## Working detail

- `.agents/skills/no-ai-slop/SKILL.md` -- the rules with worked WRONG/RIGHT
  examples.
- `.agents/skills/no-ai-slop/references/ai-writing-detection.md` -- the full
  categorized lists of banned words, phrases, and structural patterns, kept
  current against 2026 detection research.

## Operating rules

- Whenever you write or edit prose, read the no-ai-slop skill first.
- Before returning any prose, self-check it against the detection reference.
  Fix what you find.
- Apply the rules to your own output too. This file, every skill, and every
  reply obeys the rules it states.

## What changed for 2026

Detection research moved past the em-dash as the headline tell. Corpus work in
2025-2026 (WriteHuman's 80k-pair analysis, Pangram's signal study, the
TextPulse vocabulary fingerprint, Wikipedia's Signs of AI writing field guide)
shows the stronger signals are structural: hedging verbs such as "ensuring" and
"reflects", formulaic frames such as "plays a crucial role in shaping", tailing
"-ing" clauses that assert significance, "rather than" used to dodge direct
comparison, and triads padded into every claim. The rules below cover both the
classic tells and these.

The em-dash stays banned here. It is still over-represented in AI text and the
house style already forbids it.

## The rules

These are non-negotiable. Violating any of them makes the output unusable.

1. **No emdashes.** The character is banned. Use a comma, a period,
   parentheses, or restructure the sentence.

2. **No unsourced statistics.** Every number must be real and attributable.
   If you cannot point to where it comes from, do not write it. A made-up
   figure is worse than no figure.

3. **No parenthetical clarifications in headings.** Trust the reader.

4. **No intensifiers.** "Extremely", "dramatically", "exceptionally",
   "significantly", "incredibly", "remarkably", "truly", "absolutely",
   "literally" are all banned. Prove it with a fact or cut the word.

5. **No hollow statements.** Every claim must end with a concrete, verifiable
   detail. If it cannot, delete the sentence.

6. **No repeated talking points.** Say it once. Duplicates are padding.

7. **Vary structure.** Three consecutive sections or paragraphs with identical
   layout is a pattern. Break it.

8. **Reference without narrating the reference.** Do not write "as discussed
   above" or "as we will see". Make the connection and move on.

9. **No performative urgency without a reason.** "Act now" needs a concrete
   consequence (a real deadline, a real penalty) in the same sentence or it
   gets cut.

10. **No scare quotes on normal words.** Use quotation marks only for actual
    quotations from a named source.

11. **No filler phrases.** Banned: "In today's world", "It's important to
    note", "When it comes to", "At the end of the day", "In the realm of",
    "It goes without saying", "This is where X comes in", "Look no further",
    "Our team of experts".

12. **Never start a sentence with "Whether you're".**

13. **Write like a researcher, not a copywriter.** Direct, specific,
    well-grounded. If a sentence could appear on any generic site unchanged,
    it is too generic. Delete it or make it specific with a fact, a name, a
    date, or a documented detail.

14. **No synthetic enthusiasm.** No exclamation marks or cheerleading. State
    the facts. The evidence carries the weight.

15. **No weasel words.** "Helps ensure", "may be able to", "can potentially";
    either it does or it does not. Commit or cut.

16. **No narrative, dramatic, or AI-generic headings.** Headings must be
    concrete and descriptive. Do not use narrative framing ("The Right to
    Repair Trap"), thriller-style mystery ("The Hidden Cost of
    Serialization"), clickbait structure ("Why Apple Destroys Your Right to
    Repair"), or vague analytical headings ("Broader pattern", "Broader
    implications", "Wider context", "Larger trend", "Industry-wide impact").
    A heading describes what the section contains, not what it means.

17. **No fabricated case studies or scenarios.** Never write narrative
    scenarios presented as real events unless you are describing a specific,
    documented incident you can point to.

18. **No fabricated history or milestones.** Do not invent dates for events,
    launches, founding, or milestones. Every date and event must be real.

19. **No fabricated attributions.** Never claim a person, organization, or
    company said something unless it is real and verifiable. Every attributed
    quote or position must trace to a real document, transcript, public
    statement, or report.

20. **No AI transition phrases.** Banned: "Furthermore", "Moreover",
    "Notwithstanding", "That being said", "At its core", "In essence", "It is
    worth noting that", "In the landscape of", "To put it simply", "Thereby",
    "Consequently". Use plain connectors: also, and, but, however, still, so.

21. **No AI verbs.** Banned: delve, leverage, utilize, facilitate, foster,
    bolster, underscore, unveil, navigate (metaphorical), streamline,
    endeavour, ascertain, elucidate. Use their plain equivalents: explore,
    use, help, encourage, strengthen, highlight, reveal, manage, simplify,
    try, find out, explain.

22. **No academic AI tells.** Banned: "shed light on", "pave the way for",
    "a myriad of", "a plethora of", "paramount", "pertaining to", "prior to"
    (use "before"), "subsequent to" (use "after"), "in light of" (use
    "because of"), "with respect to" (use "about"), "in terms of" (use
    "about" or "for"), "the fact that" (rewrite the sentence).

23. **Quote sources accurately, and set off the long ones.** Every word inside
    quotation marks must match the source exactly. Mark any alteration with
    square brackets or paraphrase without quotes. Name the speaker and the
    medium. Set off quotes longer than about fifteen words as their own
    block with a one-sentence attribution.

24. **No research-process narration.** Report the facts you can support and
    silently omit what you cannot. Do not narrate failed searches ("could not
    be located", "was not found"), do not attach "as of [date]" to your own
    inability to find something, and do not list documents you could not
    obtain.

25. **No padding verbs.** "Ensuring", "ensures", "highlights", "supports",
    "reflects" used to drape a claim in false significance are banned. Say
    what the thing does, or delete the clause. The same ban covers "capable
    of" and "positioned to"; write "able to".

26. **No formulaic significance frames.** Banned constructions: "plays a
    crucial/critical/important/key role in", "role in shaping", "serves as a
    testament to", "a stark reminder", "is essential for" used as filler.
    Name the effect directly instead of framing its importance.

27. **No tailing "-ing" clauses that assert meaning.** Sentences that end
    with "emphasizing the significance of", "reflecting the continued
    relevance of", "highlighting the need for", "underscoring the importance
    of" bolt a vague claim of importance onto a finished fact. End the
    sentence at the fact, or write what the fact actually changed.

28. **No hedged comparisons.** "Rather than" used to soften a claim instead
    of making it is the strongest multi-word tell in 2026 corpus data. Make
    the comparison directly ("X does Y", "X costs more than Y") or cut the
    qualifier.

29. **No "not only X, but also Y".** The construction reads as generated
    emphasis. Write "X and Y" or split the thought into two sentences.

30. **No forced triads.** The rule of three is a tell when every claim,
    example list, or sentence lands in a group of three. Use two examples or
    four when that is what the evidence has. Real lists have the length the
    facts give them.

31. **No chatbot scaffolding.** Never emit "Sure!", "Great question",
    "Here's the thing:", "The bottom line:", or opener/closer boilerplate
    such as "I hope this helps" or "Let me know if you have any questions".
    Start on the content and stop when it is done.

32. **Prose over bullet lists for argument.** Bullets and numbered lists are
    for enumerations of items: steps, files, options, affected versions.
    Analysis and explanation go in paragraphs. A doc that is mostly bullets
    with bolded lead-ins reads as generated even when every line is true.

33. **No excessive boldface.** Bold is for UI elements, warnings, and the
    rare term that genuinely needs it. Bolding product names, concepts, and
    inline headers every few lines is an AI tell.

## Banned words and phrases

The full categorized lists live in
`.agents/skills/no-ai-slop/references/ai-writing-detection.md`: banned verbs,
adjectives, nouns, intensifiers, openers, transitions, closers, heading
anti-patterns, academic tells, hedging markers, structural and statistical
patterns, and per-model fingerprints. Self-check every piece of prose against
that file before returning it. If a banned word or phrase appears in your
output, the output fails.
