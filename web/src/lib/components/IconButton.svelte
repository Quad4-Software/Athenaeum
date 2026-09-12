<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";

  interface Props extends HTMLButtonAttributes {
    ariaLabel: string;
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
    ...rest
  }: Props = $props();
</script>

<button
  {type}
  class="btn btn-ghost btn-icon {className}"
  aria-label={ariaLabel}
  {title}
  {disabled}
  {onclick}
  {...rest}
>
  {@render children()}
</button>
