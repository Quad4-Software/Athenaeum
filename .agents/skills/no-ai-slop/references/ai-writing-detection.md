# AI Writing Detection

Words, phrases, punctuation patterns, structural signals, and statistical measures commonly associated with AI-generated text. Avoid these to ensure writing sounds natural and human.

Sources: Grammarly (2025), Microsoft 365 Life Hacks (2025), GPTHuman (2025), Walter Writes (2025), Textero (2025), Plagiarism Today (2025), Rolling Stone (2025), MDPI Blog (2025), isgpt.org corpus analysis (2025), ACL hedging study (2024), Wikipedia AI content detection project (2025-2026), Segmental entropy research (arxiv, 2025), WriteHuman 80,141-pair humanization corpus (April 2026), Pangram nine-signals study (2026), TextPulse vocabulary fingerprint study across six model families (2026), Kobak et al. PubMed excess-word analysis (2024), Antislop framework (arxiv 2510.15061, 2025), Measuring AI Slop taxonomy (arxiv 2509.19163, 2025), SlopDetector AI words list (2026)

## Read this first: what the 2026 data changed

Two corrections to the older sections below, based on 2025-2026 corpus studies:

1. **The em-dash is now a secondary tell.** In WriteHuman's 80,141-pair corpus (April 2026), only 18.5% of AI inputs contained one, a 2.6x gap over human text, not the near-certain marker it was in 2024. It stays banned in this repo as house style, but clean punctuation does not mean clean prose. The stronger current signals are hedging verbs, formulaic significance frames, tailing "-ing" clauses, and hedged comparisons (see the 2026 Primary Tells section).
2. **Single words are weak evidence; density and formulaic frames are the tell.** One "crucial" means little. A paragraph where every claim runs through "ensuring", "supports", or "plays a role in" is the signature. Judge clusters and sentence shapes, not isolated vocabulary.

---

## Contents
- Em Dashes: A Secondary Tell Since 2026
- 2026 Primary Tells (padding verbs, significance frames, tailing -ing clauses, hedged comparisons, not-only constructions, forced triads, chatbot scaffolding)
- Formatting Tells (markdown leakage, excessive boldface, bullets over prose)
- Overused Verbs
- Overused Adjectives
- Overused Transitions and Connectors
- Phrases That Signal AI Writing (Opening, Transitional, Concluding, Structural, Inflated Symbolism)
- Filler Words and Empty Intensifiers
- Heading Anti-Patterns
- Academic-Specific AI Tells
- Hallucinated Markup Artifacts
- Hedging and Epistemic Modality Overload
- Structural and Statistical Patterns
- Model-Family-Specific Tells
- False Positive Prevention
- How to Self-Check

## Em Dashes: A Secondary Tell Since 2026

**The em dash is still over-represented in AI text, but no longer the headline signal.** WriteHuman's April 2026 corpus (80,141 pairs) found at least one em-dash in 18.5% of AI inputs versus 7.1% of humanized text: a real 2.6x gap, far below the "everything has em-dashes" caricature. Pangram's 2026 study still measures a 10x over-representation across its corpus. Newer model generations appear to use them less.

Em dashes are longer than hyphens (-) and are used for emphasis, interruptions, or parenthetical information. While they have legitimate uses in writing, AI models overuse them, and this repo bans them outright.

### What To Do Instead
| Instead of | Use |
|------------|-----|
| The results—which were surprising—showed... | The results, which were surprising, showed... |
| This approach—unlike traditional methods—allows... | This approach, unlike traditional methods, allows... |
| The study found—as expected—that... | The study found, as expected, that... |
| Communication skills—both written and verbal—are essential | Communication skills (both written and verbal) are essential |

### Guidelines
- Use commas for most parenthetical information
- Use colons to introduce explanations or lists
- Use parentheses for supplementary information
- In this repo the em dash is banned entirely; do not use it at all

---

## 2026 Primary Tells

The strongest signals in 2025-2026 corpus data are structural: padding verbs, formulaic frames, and clause shapes. These outrank vocabulary flair and punctuation.

### Padding verbs (WriteHuman 2026, ranked by over-representation in AI inputs)

| Rank | Word | AI vs humanized |
|------|------|-----------------|
| 1 | ensuring | 4.3x |
| 2 | highlights | strong |
| 3 | supports | strong |
| 5 | broader | strong |
| 7 | ensures | strong |
| 8 | essential | strong |
| 10 | reflects | strong |
| 11 | significantly | strong |
| 12 | effectively | strong |

Also over-represented: "often", "directly", "reducing", "reduce", "strong", "rather". Humans reach for "able", "case", "order", "various" where AI writes "capable of", "positioned to", "suited for". Swap "capable of X" for "able to X", or better, state the measured capability.

Rule of thumb: a human says what the thing does. These verbs exist to make a plain fact sound considered. Cut them or replace with the concrete verb.

### Formulaic significance frames

The single most formulaic ChatGPT sentence shape in the 2026 data is "X plays a crucial/critical/important role in shaping Y". The top trigram cluster: "is essential for", "role in shaping", "crucial role in", "critical role in", "important role in", "this paper introduces", "rather than relying", "understanding of how", "aligns well with".

Banned frames:
- "plays a [adjective] role in [verb-ing]"
- "serves as a testament to"
- "a stark reminder"
- "stands as a testament to"
- "marks a shift/turning point in"
- "is essential for" used as filler (fine when naming a real dependency)

### Tailing "-ing" clauses that assert meaning

Wikipedia's Signs of AI writing guide flags the present-participle tail: a complete fact followed by a clause claiming it matters. Banned tails:
- "emphasizing the significance of"
- "reflecting the continued relevance of"
- "highlighting the need for"
- "underscoring the importance of"
- "demonstrating the value of"
- "paving the way for" (also an academic tell)

End the sentence at the fact, or write what the fact changed.

### Hedged comparisons

"Rather than" is the strongest multi-word tell in the WriteHuman corpus (17,251 occurrences in AI inputs vs 6,859 in humanized). The model uses it to soften a comparison it could state directly. Also flagged: "rather than relying", "as opposed to" used the same way, and "instead of" when it pads rather than contrasts.

### "Not only X, but also Y"

Pangram measures this construction at 3x the human rate. It is generated emphasis, not contrast. Write "X and Y" or two sentences.

### Forced triads (rule of three abuse)

Pangram: triads appear 4x more often in AI text. The tell is not using three items once; it is every claim, every example list, and every sentence resolving into a group of three. Real evidence has the length the facts give it. If you listed three adjectives because three felt complete, delete one or find the fourth real one.

### Chatbot scaffolding and sycophancy

Banned in output of any kind:
- "Sure!", "Certainly!", "Great question", "That's a great question"
- "Here's the thing:", "Think about it:", "The bottom line:", "The reality:"
- "I hope this helps", "Let me know if you have any questions", "Feel free to"
- "It's a compelling read" and similar review-bot closers

### Latinate substitutions (TextPulse 2026)

Across 60,786 human texts and rewrites by eight model configurations, the strongest AI-leaning vocabulary is formal connectives and Latinate swaps: "thereby" (13x human rate), "consequently" (11x), "utilized" (7x). Models also suppress plain words: "used" falls to one ninth of its human rate. The AI-leaning set is 44% Latinate origin against 10% for the human-leaning set, with nearly twice the syllable count. "Meticulously" is the most enriched single word measured (214x). When a plain Anglo-Saxon word exists, use it.

One caution from the same study: "delve" is not enriched in rewriting tasks, and word lists alone do not generalize across models (the BEA 2025 GPTZero-vocabulary study found near-random performance on Claude-generated text). Weight structural patterns over word lists.

---

## Formatting Tells (Pangram 2026)

Measured multipliers against human baselines, millions of documents:

| Signal | Multiplier |
|--------|-----------|
| Markdown syntax inside plain-text contexts | 12x (bold `**` 43x, `#` headers 23x, inline code 5x) |
| Stock phrases no human strings together | 12x |
| Em dash | 10x |
| Bullet points where prose belongs | 9x |
| Rule-of-three triads | 4x |
| "Not only X, but also Y" | 3x |
| Unicode characters nobody types (fancy quotes, unusual bullets) | 3x |
| Chatbot headers ("Sure! Here's...") | 2x |
| Functional emoji (checkmarks, arrows as UI in prose) | 2x |

Implications for this repo:
- In markdown docs, headings and bold are fine. The tell is density: bolding concepts, product names, and inline headers every few lines.
- Bullets are for enumerations (steps, files, options), not for analysis. Argument and explanation go in paragraphs.
- No emoji anywhere in this repo, decorative or functional.

---

## Overused Verbs

| Avoid | Use Instead |
|-------|-------------|
| delve (into) | explore, examine, investigate, look at |
| leverage | use, apply, draw on |
| optimise | improve, refine, enhance |
| utilise | use |
| facilitate | help, enable, support |
| foster | encourage, support, develop, nurture |
| bolster | strengthen, support, reinforce |
| underscore | emphasise, highlight, stress |
| unveil | reveal, show, introduce, present |
| navigate | manage, handle, work through |
| streamline | simplify, make more efficient |
| enhance | improve, strengthen |
| endeavour | try, attempt, effort |
| ascertain | find out, determine, establish |
| elucidate | explain, clarify, make clear |

---

## Overused Adjectives

| Avoid | Use Instead |
|-------|-------------|
| robust | strong, reliable, thorough, solid |
| comprehensive | complete, thorough, full, detailed |
| pivotal | key, critical, central, important |
| crucial | important, key, essential, critical |
| vital | important, essential, necessary |
| transformative | significant, important, major |
| cutting-edge | new, advanced, recent, modern |
| groundbreaking | new, original, significant |
| innovative | new, original, creative |
| seamless | smooth, easy, effortless |
| intricate | complex, detailed, complicated |
| nuanced | subtle, complex, detailed |
| multifaceted | complex, varied, diverse |
| holistic | complete, whole, comprehensive |

### Overused Metaphorical Nouns (2025-2026)
AI models use these nouns metaphorically to inject false gravitas. Literal uses are fine.

| Avoid (metaphorical) | Acceptable (literal) |
|-------|-------------|
| tapestry ("a tapestry of regulations") | tapestry (actual woven fabric) |
| symphony ("a symphony of features") | symphony (actual musical composition) |
| beacon ("a beacon of hope") | beacon (actual light or signal device) |
| realm ("in the realm of cybersecurity") | realm (actual kingdom or territory) |
| testament ("a testament to innovation") | testament (actual legal document, e.g., last will and testament) |

---

## Overused Transitions and Connectors

| Avoid | Use Instead |
|-------|-------------|
| furthermore | also, in addition, and |
| moreover | also, and, besides |
| notwithstanding | despite, even so, still |
| that being said | however, but, still |
| at its core | essentially, fundamentally, basically |
| to put it simply | in short, simply put |
| it is worth noting that | note that, importantly |
| in the realm of | in, within, regarding |
| in the landscape of | in, within |
| in today's [anything] | currently, now, today |

---

## Phrases That Signal AI Writing

### Opening Phrases to Avoid
- "In today's fast-paced world..."
- "In today's digital age..."
- "In an era of..."
- "In the ever-evolving landscape of..."
- "In the realm of..."
- "It's important to note that..."
- "Let's delve into..."
- "Imagine a world where..."

### Transitional Phrases to Avoid
- "That being said..."
- "With that in mind..."
- "It's worth mentioning that..."
- "At its core..."
- "To put it simply..."
- "In essence..."
- "This begs the question..."

### Concluding Phrases to Avoid
- "In conclusion..."
- "To sum up..."
- "By [doing X], you can [achieve Y]..."
- "In the final analysis..."
- "All things considered..."
- "At the end of the day..."

### Structural Patterns to Avoid
- "Whether you're a [X], [Y], or [Z]..." (listing three examples after "whether")
- "It's not just [X], it's also [Y]..."
- "Think of [X] as [elaborate metaphor]..."
- Starting sentences with "By" followed by a gerund: "By understanding X, you can Y..."
- Contrasting parallelisms: "It's not X. It's Y." or "It's not about X, it's about Y." More than two of these in a 500-word block is a high-confidence AI indicator.

### Inflated Symbolism Phrases (2025-2026 AI Tells)
These multi-word phrases appear hundreds of times more frequently in AI-generated text than in human baselines (corpus analysis, isgpt.org 2025):
- "provide a valuable insight" (468x more frequent in AI text)
- "left an indelible mark" (317x)
- "play a significant role in shaping" (207x)
- "an unwavering commitment" (202x)
- "open a new avenue" (174x)
- "a stark reminder" (166x)
- "gain a comprehensive understanding" (120x)
- "serves as a testament"
- "watershed moment"
- "deeply rooted"

---

## Heading Anti-Patterns

AI-generated content frequently uses narrative, dramatic, or clickbait heading structures that read like thriller chapter titles. These patterns signal low-effort AI writing even when the body text is clean. All headings must describe the section content directly and technically.

### Banned Heading Structures

| Pattern | Bad Example | Good Replacement |
|---------|-------------|------------------|
| "The [Concept] Trap" | "The Initialization Trap" | "Import vs. Initialize: DDF Metadata Destruction Risk" |
| "The [Adjective] [Noun]" drama | "The Hidden Danger" | "Firmware Corruption After Sudden Power Loss" |
| "The [Noun] [Dramatic Noun]" | "The Silent Killer" | "Gradual Bad Sector Growth on Aging Platters" |
| "Why [Action] [Dramatic Verb] [Object]" | "Why Rebuilding Destroys Everything" | "How Forced Rebuilds Overwrite Parity on Degraded Arrays" |
| "[Noun]: The [Adjective] [Noun]" | "Encryption: The Hidden Trap" | "Hardware AES-256 Encryption on WD Passport Bridge Boards" |
| "The [Noun] You [Emotion Verb]" | "The Risk You Overlook" | "Unmonitored SMART Threshold Warnings" |

### How to Self-Check Headings

1. Could this heading serve as a thriller chapter title or YouTube clickbait thumbnail? If yes, rewrite it.
2. Does the heading describe what the section contains, or does it tease it? Headings describe; they do not tease.
3. Remove "The" from the beginning of any heading and check if it still uses a dramatic noun pairing. If so, rewrite.
4. A good heading reads like an entry in a technical manual index: specific, descriptive, and boring to non-specialists.

---

## Filler Words and Empty Intensifiers

These words often add nothing to meaning. Remove them or find specific alternatives:

- absolutely
- actually
- basically
- certainly
- clearly
- definitely
- essentially
- extremely
- fundamentally
- incredibly
- interestingly
- naturally
- obviously
- quite
- really
- significantly
- simply
- surely
- truly
- ultimately
- undoubtedly
- very

---

## Academic-Specific AI Tells

| Avoid | Use Instead |
|-------|-------------|
| shed light on | clarify, explain, reveal |
| pave the way for | enable, allow, make possible |
| a myriad of | many, numerous, various |
| a plethora of | many, numerous, several |
| paramount | very important, essential, critical |
| pertaining to | about, regarding, concerning |
| prior to | before |
| subsequent to | after |
| in light of | because of, given, considering |
| with respect to | about, regarding, for |
| in terms of | regarding, for, about |
| the fact that | that (or rewrite sentence) |

---

## Hallucinated Markup Artifacts

When AI generates wikitext, it sometimes hallucinates citation markup from its training data. These are 100% confidence indicators of unedited AI output:

| Artifact | Origin |
|----------|--------|
| `oaicite` | OpenAI ChatGPT citation placeholder |
| `contentReference` | OpenAI internal reference tag |
| `grok_card` | xAI Grok citation tag |
| `attributableIndex` | AI attribution tracking artifact |
| `turn0search0` | ChatGPT search result placeholder |

Any occurrence of these strings in wikitext means the text was pasted from an AI tool without editing. Zero tolerance.

---

## Hedging and Epistemic Modality Overload

AI models hedge 4-7x more than human writers (ACL 2024 study, 12,000 technical documents). Because models are trained to avoid stating hallucinations as facts, they default to blanket hedging even for established facts.

### Hedging Markers
**Epistemic modals** (45% of AI hedges): may, might, could, potentially
**Cognitive verbs** (25%): I think, I believe, it seems, it appears
**Adverbs of limitation** (20%): probably, generally, usually, arguably, likely
**Explicit uncertainty markers**: unclear, remains to be seen, further research is needed

### Thresholds
- **Per-paragraph:** More than 3 hedging instances in a single paragraph warrants scrutiny
- **Per-1000-words:** More than 8 hedging markers per 1,000 words in declarative sections (Background, History, Timeline) indicates AI generation. These sections state established facts.
- **Appropriate hedging:** Sections discussing pending legislation, ongoing litigation, or genuinely disputed facts should hedge. Do not flag hedging in those contexts.

### AI Hedging Phrases to Flag
- "It is worth noting that..."
- "It should be noted that..."
- "One could argue that..."
- "While X, Y remains..."
- "Though precise thresholds can vary depending on..."
- "It is widely acknowledged that..."

### Human vs. AI Hedging
Humans hedge contextually, grounding uncertainty in specific evidence: "The FTC's 2024 enforcement data suggests a 12% increase." AI hedges with blanket qualifiers on established facts: "It is widely acknowledged that repair restrictions may potentially impact consumers."

---

## Structural and Statistical Patterns

Beyond lexical tells, AI text exhibits measurable structural uniformity that human writing does not.

### Paragraph Length Uniformity
AI aims for visual symmetry. Paragraphs tend toward identical sentence counts (typically 3-4 sentences each). Human writing varies paragraph length based on sub-topic complexity.
- **Threshold:** If all paragraphs in a section are within 15% of each other in word count, the section is likely AI-generated.
- **Exception:** Bulleted lists, tables, and template fields are structurally uniform by design.

### Sentence Length Uniformity (Burstiness)
Human writing alternates between short, punchy sentences and long, clause-heavy ones. AI sentences cluster uniformly around 15-20 words.
- **Threshold:** If a 500-word block contains no sentences under 8 words or over 30 words, it lacks human burstiness.
- **Human baseline:** Human text exhibits 3+ distinct syntactic patterns per 100 words. AI text shows 1.5 or fewer.

### Transition Density
AI over-relies on transition words and adverbial clauses to maintain flow between paragraphs.
- **Threshold:** If more than 30% of paragraphs in an article begin with a transition word or adverbial clause, the text is structurally artificial.

### Opening-Word Repetition
Three or more consecutive paragraphs starting with the same word or phrase pattern indicates mechanical generation. Vary opening words.

### Segmental Entropy
AI maintains flat stylistic consistency from introduction through conclusion. Human writers naturally vary pacing, complexity, and sentence structure between sections.
- **Threshold:** Calculate sentence length variance separately for the introduction, body, and conclusion. If variance differs by less than 10% across all three segments, the text was likely generated as a single pass by AI.
- **Why this matters:** Human introductions tend to be tighter and more declarative. Human body sections are denser with longer sentences. Human conclusions shift register. AI maintains a monotone throughout.

### Contrasting Parallelism Overuse
2025-era models overuse sequential contrasting structures to simulate punchy emphasis:
- "It's not X, it's Y."
- "It's not about X, it's about Y."
- "The issue isn't X. The issue is Y."
- **Threshold:** More than two contrasting parallelisms in a 500-word block.

---

## Model-Family-Specific Tells

Different model families produce distinct stylistic fingerprints, and the fingerprints shift with each generation. Treat these as priors, not proof: the BEA 2025 study found vocabulary lists tuned on ChatGPT text perform near-random on Claude-generated text.

### GPT family (OpenAI)
- Heavy use of bullet-point formatting and structured lists
- Staccato short-sentence contrasting: "It's not X. It's Y." used to simulate punchy copy
- Rhetorical colon abuse: "Here's the thing:", "Think about it:", "The bottom line:", "The reality:"
- Over-structures arguments into numbered steps
- 2026-era GPT text leans hardest on the padding-verb cluster: "ensuring", "ensures", "highlights", "supports", "reflects"

### Claude family (Anthropic)
- Better sentence length variation than GPT, but still exhibits flat segmental entropy
- Overly polite and conciliatory transitions: "It's worth considering that", "To be fair", "That said"
- Leans toward poetic and metaphorical prose with words like "nuanced," "complexities"
- Loses thread in long documents and resorts to increasingly generic transitions
- Tends toward diplomatic hedging even when stating documented facts

### Common Across All Models
- Uniform paragraph lengths
- Predictable section ordering (Background > Details > Impact > Response)
- Citation clustering at paragraph ends rather than distributed throughout sentences
- Excessive boldface on concepts, product names, and inline headers
- Formulaic significance frames ("plays a role in shaping") and tailing "-ing" clauses

### What is not a tell
- **Semicolons**: 1.18 vs 1.10 per 1,000 words (AI vs humanized), effectively identical
- **"Delve" in rewritten text**: enriched in from-scratch generation, not in rewrites
- **Sentence length alone**: AI input averaged 22.9 words per sentence vs 23.3 humanized; variance within a document matters, not the mean
- **Single occurrences of any flagged word**: density and clustering carry the signal

---

## False Positive Prevention

### Exclusion Zones
Lexical scans must NOT flag text inside:
- Direct quotes (`"..."`) from cited sources
- Titles, names, and other verbatim values taken from a source
- Code, configuration, or markup that is being shown as an example

### Context-Aware Severity
If a banned word appears immediately adjacent to specific named entities (proper nouns, statute numbers, dates, dollar amounts), it is more likely being used with technical meaning than as AI filler. Reduce flag severity.
- **Higher severity:** "a comprehensive examination of the issues" (abstract nouns, no specifics)
- **Lower severity:** "comprehensive audit by the FTC in 2024" (specific entity, specific date)

### Metaphorical vs. Literal Distinction
These words require bigram context checking. Only flag metaphorical uses:
- ecosystem: "Apple's software ecosystem" (OK) vs. "the repair ecosystem" (flag)
- landscape: "Arizona landscape" (OK) vs. "the regulatory landscape" (flag)
- navigate: "navigate the website" (OK) vs. "navigate the regulatory process" (flag)
- tapestry: "medieval tapestry" (OK) vs. "a tapestry of regulations" (flag)
- symphony: "Beethoven's symphony" (OK) vs. "a symphony of features" (flag)
- beacon: "lighthouse beacon" (OK) vs. "a beacon of hope" (flag)
- testament: "last will and testament" (OK) vs. "a testament to innovation" (flag)

---

## How to Self-Check

1. Read your text aloud. If phrases sound unnatural in speech, revise them
2. Ask: "Would I say this in a conversation with a colleague?"
3. Check for repetitive sentence structures
4. Look for clusters of the words listed above. One flagged word is weak evidence; a paragraph built on "ensuring", "supports", and "plays a role in" is not.
5. Ensure varied sentence lengths (not all similar length)
6. Verify each intensifier adds genuine meaning
7. Count hedging markers per paragraph. More than 3 in a single paragraph is a red flag.
8. Check paragraph word counts within each section. If they are all similar, vary them.
9. Search for hallucinated markup: `oaicite`, `contentReference`, `turn0search0`, `grok_card`
10. Check if your introduction, body, and conclusion have different pacing and sentence complexity
11. Scan for the 2026 tells: tailing "-ing" significance clauses, "rather than" hedged comparisons, "not only X but also Y", forced triads, chatbot openers and closers
12. Count boldface spans and bullet lists. Prose carries argument; lists carry enumerations.
