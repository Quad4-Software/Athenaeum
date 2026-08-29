<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    type?: "button" | "submit" | "reset";
    ariaLabel: string;
    title?: string;
    class?: string;
    disabled?: boolean;
    onclick?: (event: MouseEvent) => void;
    children: Snippet;
  }

  let {
    type = "button",
    ariaLabel,
    title,
    class: className = "",
    disabled = false,
    onclick,
    children,
  }: Props = $props();
</script>

<button
  {type}
  class="btn btn-ghost btn-icon {className}"
  aria-label={ariaLabel}
  {title}
  {disabled}
  {onclick}
>
  {@render children()}
</button>
