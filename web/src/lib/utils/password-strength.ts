export interface PasswordPolicy {
  minLength: number;
  longLength: number;
  minKinds: number;
  requireLower: boolean;
  requireUpper: boolean;
  requireDigit: boolean;
  requireSymbol: boolean;
}

export interface PasswordRequirement {
  id: string;
  met: boolean;
}

export interface PasswordStrength {
  score: number;
  label: "weak" | "fair" | "good" | "strong";
  valid: boolean;
  issues?: string[];
  requirements: PasswordRequirement[];
}

export interface AuditEntry {
  id: number;
  actorId: number;
  actorName: string;
  targetUserId?: number;
  targetName?: string;
  action: string;
  details?: string;
  ip?: string;
  createdAt: string;
}

export interface AuditPage {
  items: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}

export const DEFAULT_PASSWORD_POLICY: PasswordPolicy = {
  minLength: 8,
  longLength: 12,
  minKinds: 3,
  requireLower: false,
  requireUpper: false,
  requireDigit: false,
  requireSymbol: false,
};

const LOWER = /[a-z]/;
const UPPER = /[A-Z]/;
const DIGIT = /[0-9]/;
const SYMBOL = /[^a-zA-Z0-9]/;

export function normalizePasswordPolicy(policy?: PasswordPolicy | null): PasswordPolicy {
  const p = { ...DEFAULT_PASSWORD_POLICY, ...policy };
  if (p.minLength < 4) p.minLength = 4;
  if (p.minLength > 128) p.minLength = 128;
  if (p.longLength < 0) p.longLength = 0;
  if (p.longLength > 256) p.longLength = 256;
  if (p.longLength > 0 && p.longLength < p.minLength) p.longLength = p.minLength;
  if (p.minKinds < 0) p.minKinds = 0;
  if (p.minKinds > 4) p.minKinds = 4;
  return p;
}

export function scorePassword(password: string, policy?: PasswordPolicy | null): PasswordStrength {
  const p = normalizePasswordPolicy(policy);
  const issues: string[] = [];

  const hasLower = LOWER.test(password);
  const hasUpper = UPPER.test(password);
  const hasDigit = DIGIT.test(password);
  const hasSymbol = SYMBOL.test(password);

  let kinds = 0;
  if (hasLower) kinds++;
  if (hasUpper) kinds++;
  if (hasDigit) kinds++;
  if (hasSymbol) kinds++;

  const lengthOK = password.length >= p.minLength;
  let diversityOK = true;
  if (p.minKinds > 0) {
    const longOK = p.longLength > 0 && password.length >= p.longLength;
    diversityOK = longOK || kinds >= p.minKinds;
  }

  const requirements: PasswordRequirement[] = [{ id: "minLength", met: lengthOK }];
  if (p.requireLower) requirements.push({ id: "requireLower", met: hasLower });
  if (p.requireUpper) requirements.push({ id: "requireUpper", met: hasUpper });
  if (p.requireDigit) requirements.push({ id: "requireDigit", met: hasDigit });
  if (p.requireSymbol) requirements.push({ id: "requireSymbol", met: hasSymbol });
  if (p.minKinds > 0) requirements.push({ id: "diversity", met: diversityOK });

  if (!lengthOK) issues.push(`at least ${p.minLength} characters`);
  if (p.requireLower && !hasLower) issues.push("include a lowercase letter");
  if (p.requireUpper && !hasUpper) issues.push("include an uppercase letter");
  if (p.requireDigit && !hasDigit) issues.push("include a digit");
  if (p.requireSymbol && !hasSymbol) issues.push("include a symbol");
  if (p.minKinds > 0 && !diversityOK) {
    if (p.longLength > 0) {
      issues.push(`use ${p.longLength}+ characters or mix upper, lower, digits, and symbols`);
    } else {
      issues.push(`use at least ${p.minKinds} character types`);
    }
  }

  let score = 0;
  if (lengthOK) score++;
  if (p.longLength > 0 && password.length >= p.longLength) score++;
  else if (p.longLength === 0 && lengthOK && password.length >= p.minLength + 4) score++;
  if (kinds >= 3) score++;
  if (kinds >= 4 && password.length >= Math.max(10, p.minLength)) score++;

  let label: PasswordStrength["label"] = "weak";
  if (score === 2) label = "fair";
  else if (score === 3) label = "good";
  else if (score >= 4) label = "strong";

  const valid =
    lengthOK &&
    diversityOK &&
    (!p.requireLower || hasLower) &&
    (!p.requireUpper || hasUpper) &&
    (!p.requireDigit || hasDigit) &&
    (!p.requireSymbol || hasSymbol);

  return { score, label, valid, issues: issues.length ? issues : undefined, requirements };
}

export function strengthColor(label: PasswordStrength["label"]): string {
  switch (label) {
    case "weak":
      return "var(--color-danger)";
    case "fair":
      return "var(--color-warning, #d97706)";
    case "good":
      return "var(--color-primary)";
    default:
      return "var(--color-success, #16a34a)";
  }
}
