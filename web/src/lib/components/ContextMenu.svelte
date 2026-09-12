<script lang="ts">
  import { DropdownMenu } from "bits-ui";
  import MenuItems from "./MenuItems.svelte";
  import type { MenuItem } from "./menu";

  interface Props {
    open?: boolean;
    x?: number;
    y?: number;
    title?: string;
    items?: MenuItem[];
    onclose?: () => void;
  }

  let { open = $bindable(false), x = 0, y = 0, title, items = [], onclose }: Props = $props();

  const anchor = $derived({
    getBoundingClientRect: () => new DOMRect(x, y, 0, 0),
  });

  function onOpenChange(next: boolean) {
    if (!next) onclose?.();
  }
</script>

<DropdownMenu.Root bind:open {onOpenChange}>
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class="menu-panel context-panel"
      customAnchor={anchor}
      side="bottom"
      align="start"
      sideOffset={0}
      collisionPadding={8}
      onCloseAutoFocus={(e) => e.preventDefault()}
    >
      <MenuItems {title} {items} />
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.context-panel) {
    animation: context-in 140ms ease-out;
    transform-origin: top left;
  }

  @keyframes context-in {
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
