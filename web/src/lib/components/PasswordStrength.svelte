<script lang="ts">
  import { i18n } from "$lib/stores/i18n.svelte";
  import {
    DEFAULT_PASSWORD_POLICY,
    normalizePasswordPolicy,
    scorePassword,
    strengthColor,
    type PasswordPolicy,
  } from "$lib/utils/password-strength";

  interface Props {
    password: string;
    policy?: PasswordPolicy | null;
    showMeter?: boolean;
    showRequirements?: boolean;
  }

  let { password, policy = null, showMeter = true, showRequirements = true }: Props = $props();

  const resolvedPolicy = $derived(normalizePasswordPolicy(policy ?? DEFAULT_PASSWORD_POLICY));
  const strength = $derived(scorePassword(password, resolvedPolicy));

  function requirementLabel(id: string): string {
    switch (id) {
      case "minLength":
        return i18n.t("auth.passwordReqMinLength", { n: resolvedPolicy.minLength });
      case "requireLower":
        return i18n.t("auth.passwordReqLower");
      case "requireUpper":
        return i18n.t("auth.passwordReqUpper");
      case "requireDigit":
        return i18n.t("auth.passwordReqDigit");
      case "requireSymbol":
        return i18n.t("auth.passwordReqSymbol");
      case "diversity":
        return resolvedPolicy.longLength > 0
          ? i18n.t("auth.passwordReqDiversityOrLong", { n: resolvedPolicy.longLength })
          : i18n.t("auth.passwordReqDiversity", { n: resolvedPolicy.minKinds });
      default:
        return id;
    }
  }
</script>

{#if password.length > 0 || showRequirements}
  <div class="mt-2 space-y-2" aria-live="polite">
    {#if showMeter && password.length > 0}
      <div class="space-y-1">
        <div class="flex gap-1">
          {#each [0, 1, 2, 3] as i (i)}
            <div
              class="h-1 flex-1 rounded-full bg-border transition-colors"
              style:background={i < strength.score ? strengthColor(strength.label) : undefined}
            ></div>
          {/each}
        </div>
        <p class="text-xs capitalize text-muted">
          {i18n.t(`auth.passwordStrength.${strength.label}`)}
        </p>
      </div>
    {/if}

    {#if showRequirements}
      <ul class="space-y-1">
        {#each strength.requirements as req (req.id)}
          <li
            class={[
              "flex items-start gap-2 text-xs",
              req.met ? "text-success" : password.length > 0 ? "text-danger" : "text-muted",
            ]}
          >
            <span
              class={[
                "mt-0.5 inline-block size-2.5 shrink-0 rounded-full border",
                req.met ? "border-transparent bg-success" : "border-current bg-transparent",
              ]}
              aria-hidden="true"
            ></span>
            <span>{requirementLabel(req.id)}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}
