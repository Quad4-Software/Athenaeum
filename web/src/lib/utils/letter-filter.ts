/**
 * Alphabet strip for the library browse grid. "#" matches titles that do not
 * start with a letter; the rest filter by first letter of the title.
 */
export const LETTER_FILTER_OPTIONS = [
  "A",
  "B",
  "C",
  "D",
  "E",
  "F",
  "G",
  "H",
  "I",
  "J",
  "K",
  "L",
  "M",
  "N",
  "O",
  "P",
  "Q",
  "R",
  "S",
  "T",
  "U",
  "V",
  "W",
  "X",
  "Y",
  "Z",
  "#",
] as const;

/** Picking the active letter clears the filter; anything else selects it. */
export function nextLetterFilter(current: string, pick: string): string {
  return current === pick ? "" : pick;
}
