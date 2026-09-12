<script lang="ts">
  import type { Snippet } from "svelte";
  import { DropdownMenu } from "bits-ui";
  import MenuItems from "./MenuItems.svelte";
  import type { MenuItem } from "./menu";

  interface Props {
    open?: boolean;
    side?: "top" | "bottom";
    align?: "center" | "start" | "end";
    minWidth?: number;
    class?: string;
    title?: string;
    items?: MenuItem[];
    trigger: Snippet<[props: Record<string, unknown>]>;
    children?: Snippet;
    onclose?: () => void;
  }

  let {
    open = $bindable(false),
    side = "bottom",
    align = "center",
    minWidth = 0,
    class: className = "",
    title,
    items,
    trigger,
    children,
    onclose,
  }: Props = $props();

  function onOpenChange(next: boolean) {
    if (!next) onclose?.();
  }
</script>

<DropdownMenu.Root bind:open {onOpenChange}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      {@render trigger(props)}
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class="menu-panel dropdown-panel {className}"
      {side}
      {align}
      sideOffset={6}
      collisionPadding={8}
      style={minWidth > 0 ? `min-width:${minWidth}px` : undefined}
    >
      {#if items}
        <MenuItems {title} {items} />
      {/if}
      {@render children?.()}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.dropdown-panel) {
    animation: dropdown-in 120ms ease-out;
  }

  @keyframes dropdown-in {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
