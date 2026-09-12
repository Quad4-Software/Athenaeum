---
name: no-ai-slop
description: "Rules and worked examples for writing prose that does not read like AI-generated slop, updated for 2026 detection research. Consult before writing or editing any prose."
---

# No AI Slop

The full rule list lives in `.agents/references/no-ai-slop-rules.md` (rules 1 through 33). This skill turns the rules that have worked examples into actionable guidance: each shows a WRONG version (the slop) and a RIGHT version (the fix). The pattern behind every fix is the same: replace the vague claim with a specific, checkable fact.

## Rule 1: No emdashes

The character is banned. Use a comma, a period, parentheses, or restructure.

- WRONG: "The policy -- which affected millions -- was later reversed."
- RIGHT: "The policy affected millions of devices. The company reversed it in December 2017."

2026 note: the em-dash is now a secondary tell (present in 18.5% of AI text in WriteHuman's April 2026 corpus, down from near-universal use in 2024-era models). It stays banned here as house style, but do not treat its absence as proof text is clean. The stronger current tells are rules 25 through 30 below.

## Rule 4: No intensifiers

"Significantly", "dramatically", "extremely" and their kin are placeholders for evidence. Replace the word with the number it was standing in for.

- WRONG: "The pricing was significantly higher than the cost of the part."
- RIGHT: "They charged $1,200 for a repair that needed a $5 chip."

## Rule 5: No hollow statements

A sentence that asserts importance without a detail says nothing. End every claim on a concrete fact.

- WRONG: "This practice has had a significant impact on people."
- RIGHT: "The company replaced 11 million batteries in 2018, against the 1 to 2 million it had expected."

## Rule 7: No structural slop (repetitive layouts)

Three sections built from the same template read as machine output, even when each fact is true. Vary paragraph count, sentence rhythm, and how each section opens.

- WRONG (three sections, identical shape):
  ```
  In [year], [party] did [thing]. This affected [number] people. [Party] responded by [action].
  In [year], [party] did [thing]. This affected [number] people. [Party] responded by [action].
  In [year], [party] did [thing]. This affected [number] people. [Party] responded by [action].
  ```
- RIGHT (vary the shape):
  ```
  Section one: a detailed narrative with timeline and context across two paragraphs.
  Section two: a two-sentence summary, because the event is thinly documented.
  Section three: opens with the party's stated justification, then the contradicting evidence.
  ```

## Rule 11: No filler phrases

"In today's world", "It's important to note", "When it comes to" add length, not meaning. Open on the fact.

- WRONG: "In today's world, planned obsolescence affects many devices."
- RIGHT: "Apple, Samsung, and Google have each faced lawsuits alleging planned obsolescence."

## Rule 13: Write like a researcher, not a copywriter

If a sentence could sit on any advocacy or marketing site without changing a word, it is generic. Anchor it to something checkable.

- WRONG: "People deserve the right to repair their own devices."
- RIGHT: "The FTC voted 5-0 in July 2021 to step up enforcement against illegal repair restrictions."

## Rule 15: No weasel words

"May potentially", "can help to", "might be able to" hedge a claim into meaninglessness. Either the thing happens or it does not. Say which.

- WRONG: "Serialization may potentially prevent independent repair."
- RIGHT: "Replacing an iPhone 15 camera module without the manufacturer's calibration software disables optical image stabilization."

## Rule 16: No dramatic headings

A heading names what the section holds. It does not tease, dramatize, or abstract.

- WRONG: "The Hidden Cost of Planned Obsolescence"
- RIGHT: "Economic impact of shortened product lifespans"

## Rule 19: No fabricated attributions

Never put a position in a named person's mouth from inference. State only what they actually did or said, with the real source.

- WRONG: "Senator Smith has argued that the right to repair is essential."
- RIGHT: "Senator Smith co-sponsored the Fair Repair Act in January 2024."

## Rule 25: No padding verbs

"Ensuring", "ensures", "highlights", "supports", "reflects" are the strongest single-word AI tells in 2026 corpus data ("ensuring" is over-represented 4.3x). The model reaches for them when it pads an idea to sound considered. Say what the thing does, or cut the clause.

- WRONG: "The retry logic ensures reliability while highlighting the system's resilience."
- RIGHT: "Failed requests retry three times with exponential backoff, then surface as an error in the UI."

- WRONG: "The library is capable of handling large collections."
- RIGHT: "The library indexes 50,000 EPUBs without a noticeable delay." (Or, when you lack a number: "able to handle" is still weak; find the fact or cut the claim.)

## Rule 26: No formulaic significance frames

"X plays a crucial role in shaping Y" is, statistically, the single most formulaic sentence shape ChatGPT produces (top trigram cluster in the WriteHuman corpus: "role in shaping", "crucial role in", "critical role in", "important role in"). Banned along with "serves as a testament to", "a stark reminder", and filler uses of "is essential for".

- WRONG: "Local-first storage plays a crucial role in shaping the user experience."
- RIGHT: "The library lives on your disk. It works when the internet is down."

## Rule 27: No tailing "-ing" clauses that assert meaning

Wikipedia's AI cleanup project flags the present-participle tail as a reliable tell: a finished fact with a clause bolted on to claim it matters. End the sentence at the fact, or write what the fact changed.

- WRONG: "The bill passed the state senate in March, reflecting the continued relevance of repair rights."
- RIGHT: "The bill passed the state senate 31-4 in March. Manufacturers must now sell parts to independent shops in that state."

## Rule 28: No hedged comparisons

"Rather than" is the strongest multi-word tell in the 2026 data (17,251 occurrences in AI inputs against 6,859 in humanized outputs). The model uses it to hedge a comparison it could just make.

- WRONG: "The scanner favors accuracy rather than speed."
- RIGHT: "The scanner parses every file completely before indexing. A full scan of a large library takes minutes, not seconds."

## Rule 29: No "not only X, but also Y"

Generated emphasis. Write "X and Y" or split the thought.

- WRONG: "The server not only indexes your books, but also streams them to any device."
- RIGHT: "The server indexes your books and streams them to any device."

## Rule 30: No forced triads

When every claim arrives in a group of three, the grouping came from the model, not the material.

- WRONG: "The reader is fast, reliable, and intuitive. Setup is simple, secure, and seamless."
- RIGHT: "The reader opens books in under a second and remembers your position across devices. Setup is one command."

## Root-cause differentiation

When you contrast two things, name the concrete difference that separates them. Do not assert that one is exempt, newer, better, or unaffected without saying what specifically makes it so.

- WRONG: "2020+ Leaf models are unaffected and use the MyNISSAN app instead."
- RIGHT: "2020+ Leaf models shipped with 4G/LTE telematics units connected to a newer cloud platform, replacing the 2G/3G units in earlier models. Those vehicles use the MyNISSAN app, which talks to a different backend."

Whenever you say A differs from B, name the part, the version, the date, the mechanism, or the supply-chain change that makes the difference real. If you do not have that detail, do not imply the difference exists.

## Self-check before returning text

Run this pass on every piece of prose before you hand it back. The full banned lists are in `references/ai-writing-detection.md`; check against them directly.

1. Search for the emdash character. Remove every one (Rule 1).
2. Scan for banned verbs (delve, leverage, utilize, foster, bolster, underscore, unveil, streamline) and replace with plain equivalents.
3. Scan for banned adjectives and intensifiers (robust, comprehensive, pivotal, seamless, significantly, extremely, truly) and cut or replace.
4. Scan for banned transitions and openers (Furthermore, Moreover, That being said, In today's world, It's worth noting that, Thereby, Consequently).
5. Scan for the 2026 tells: padding verbs (ensuring, highlights, supports, reflects), "plays a ... role in", tailing "-ing" clauses of significance, hedged "rather than" comparisons, "not only X but also Y", forced triads.
6. Check every number: is it real and attributable? If not, cut it (Rule 2).
7. Check every sentence ends on a concrete detail, not an assertion of importance (Rule 5).
8. Check headings: does each name the content (Rule 16)?
9. Check for repeated points and repeated section shapes (Rules 6, 7).
10. Count hedging markers per paragraph. More than three is a red flag.
11. Count boldface spans. If you bolded concepts or product names as decoration, unbold them (Rule 33).
12. Read it aloud. If a phrase would sound unnatural to a colleague, rewrite it.
