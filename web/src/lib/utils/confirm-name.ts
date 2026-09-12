/**
 * Gate for type-to-confirm destructive actions. The typed text must equal the
 * expected name exactly: case-sensitive, no trimming.
 */
export function confirmNameMatches(typed: string, expected: string): boolean {
  return typed === expected;
}
