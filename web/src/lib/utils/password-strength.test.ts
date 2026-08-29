import { describe, expect, it } from "vitest";
import { scorePassword, type PasswordPolicy } from "./password-strength";

describe("scorePassword", () => {
  it("rejects short passwords", () => {
    expect(scorePassword("abc").valid).toBe(false);
  });

  it("accepts long simple passwords", () => {
    expect(scorePassword("longpassword").valid).toBe(true);
  });

  it("accepts mixed shorter passwords", () => {
    expect(scorePassword("Aa1!aaaa").valid).toBe(true);
  });

  it("honors strict require flags", () => {
    const policy: PasswordPolicy = {
      minLength: 10,
      longLength: 0,
      minKinds: 0,
      requireLower: true,
      requireUpper: true,
      requireDigit: true,
      requireSymbol: true,
    };
    expect(scorePassword("Aa1!aaaaaa", policy).valid).toBe(true);
    expect(scorePassword("aaaaaaaaaa", policy).valid).toBe(false);
    expect(
      scorePassword("aaaaaaaaaa", policy).requirements.some(
        (r) => r.id === "requireUpper" && !r.met,
      ),
    ).toBe(true);
  });
});
