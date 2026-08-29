<script lang="ts">
  import { Loader2 } from "@lucide/svelte";
  import type { Snippet } from "svelte";

  type Variant = "primary" | "ghost";
  type Size = "sm" | "md";

  interface Props {
    type?: "button" | "submit" | "reset";
    variant?: Variant;
    size?: Size;
    loading?: boolean;
    disabled?: boolean;
    class?: string;
    onclick?: (event: MouseEvent) => void;
    children: Snippet;
  }

  let {
    type = "button",
    variant = "primary",
    size = "md",
    loading = false,
    disabled = false,
    class: className = "",
    onclick,
    children,
  }: Props = $props();

  let isDisabled = $derived(disabled || loading);
</script>

<button
  {type}
  class="btn {variant === 'primary' ? 'btn-primary' : 'btn-ghost'} {size === 'sm'
    ? 'btn-sm'
    : ''} {className}"
  class:btn-loading={loading}
  disabled={isDisabled}
  aria-busy={loading}
  {onclick}
>
  {#if loading}
    <Loader2 size={size === "sm" ? 14 : 16} class="btn-spinner" />
  {/if}
  {@render children()}
</button>

<style>
  .btn-sm {
    padding: 0.375rem 0.625rem;
    font-size: 0.75rem;
  }

  .btn-loading {
    position: relative;
  }

  .btn-spinner {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
