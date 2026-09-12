<script lang="ts">
  import type { Snippet } from "svelte";
  import { Popover } from "bits-ui";

  interface Props {
    open?: boolean;
    placement?: "top" | "bottom";
    align?: "center" | "start" | "end";
    minWidth?: number;
    trigger: Snippet<[props: Record<string, unknown>]>;
    children: Snippet;
    onclose?: () => void;
    onOpenAutoFocus?: (event: Event) => void;
  }

  let {
    open = $bindable(false),
    placement = "bottom",
    align = "center",
    minWidth = 0,
    trigger,
    children,
    onclose,
    onOpenAutoFocus,
  }: Props = $props();

  function onOpenChange(next: boolean) {
    if (!next) onclose?.();
  }
</script>

<Popover.Root bind:open {onOpenChange}>
  <Popover.Trigger>
    {#snippet child({ props })}
      {@render trigger(props)}
    {/snippet}
  </Popover.Trigger>
  <Popover.Portal>
    <Popover.Content
      class="popover-panel"
      side={placement}
      {align}
      sideOffset={8}
      collisionPadding={8}
      trapFocus={false}
      style={minWidth > 0 ? `min-width:${minWidth}px` : undefined}
      {onOpenAutoFocus}
    >
      {@render children()}
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>

<style>
  :global(.popover-panel) {
    width: max-content;
    max-width: min(22rem, calc(100vw - 1rem));
    max-height: min(70vh, 28rem);
    overflow: auto;
    border-radius: var(--radius-card);
    background: var(--color-surface);
    color: var(--color-fg);
    box-shadow:
      0 1px 0 color-mix(in oklch, var(--color-fg) 6%, transparent),
      0 16px 40px -16px rgb(0 0 0 / 0.55);
    border: 1px solid var(--color-border);
    padding: 0.5rem;
    animation: popover-in 140ms ease-out;
  }

  @keyframes popover-in {
    from {
      opacity: 0;
      transform: translateY(-6px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
