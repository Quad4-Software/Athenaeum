<script lang="ts">
  import type { Snippet } from "svelte";
  import { Tooltip } from "bits-ui";

  interface Props {
    content: string;
    side?: "top" | "bottom" | "left" | "right";
    delayDuration?: number;
    children: Snippet<[props: Record<string, unknown>]>;
  }

  let { content, side = "top", delayDuration = 350, children }: Props = $props();
</script>

<Tooltip.Root {delayDuration}>
  <Tooltip.Trigger>
    {#snippet child({ props })}
      {@render children(props)}
    {/snippet}
  </Tooltip.Trigger>
  <Tooltip.Portal>
    <Tooltip.Content class="tooltip-panel" {side} sideOffset={6} collisionPadding={8}>
      {content}
    </Tooltip.Content>
  </Tooltip.Portal>
</Tooltip.Root>

<style>
  :global(.tooltip-panel) {
    z-index: 90;
    max-width: 16rem;
    border-radius: 0.5rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    color: var(--color-fg);
    padding: 0.3rem 0.55rem;
    font-size: 0.75rem;
    box-shadow: var(--shadow);
    animation: tooltip-in 120ms ease-out;
  }

  @keyframes tooltip-in {
    from {
      opacity: 0;
      transform: scale(0.96);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
