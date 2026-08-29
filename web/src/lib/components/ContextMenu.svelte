<script lang="ts">
  import { tick } from "svelte";
  import { scale } from "svelte/transition";
  import MenuList, { type MenuItem } from "./MenuList.svelte";

  interface Props {
    open?: boolean;
    x?: number;
    y?: number;
    title?: string;
    items?: MenuItem[];
    onclose?: () => void;
  }

  let { open = false, x = 0, y = 0, title, items = [], onclose }: Props = $props();

  let panelEl = $state<HTMLDivElement>();
  let panelStyle = $state("");

  function close() {
    onclose?.();
  }

  function onContextMenu(event: MouseEvent) {
    event.preventDefault();
  }

  $effect(() => {
    if (!open) return;

    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") close();
    };
    const onScroll = () => close();
    const onResize = () => close();

    window.addEventListener("keydown", onKey);
    window.addEventListener("scroll", onScroll, true);
    window.addEventListener("resize", onResize);

    void tick().then(() => {
      if (!panelEl) return;
      const margin = 8;
      const rect = panelEl.getBoundingClientRect();
      const vw = window.innerWidth;
      const vh = window.innerHeight;
      let left = x;
      let top = y;
      if (left + rect.width > vw - margin) left = vw - rect.width - margin;
      if (top + rect.height > vh - margin) top = vh - rect.height - margin;
      left = Math.max(margin, left);
      top = Math.max(margin, top);
      panelStyle = `top:${top}px;left:${left}px;`;
    });

    return () => {
      window.removeEventListener("keydown", onKey);
      window.removeEventListener("scroll", onScroll, true);
      window.removeEventListener("resize", onResize);
    };
  });
</script>

{#if open}
  <div
    class="context-backdrop"
    role="presentation"
    oncontextmenu={onContextMenu}
    onclick={close}
  ></div>
  <div
    bind:this={panelEl}
    class="context-panel"
    style={panelStyle}
    in:scale={{ duration: 140, start: 0.96, opacity: 0 }}
  >
    <MenuList {title} {items} />
  </div>
{/if}

<style>
  .context-backdrop {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: transparent;
  }

  .context-panel {
    position: fixed;
    z-index: 70;
    min-width: 12rem;
    max-width: min(18rem, calc(100vw - 1rem));
    border-radius: var(--radius-card);
    border: 1px solid var(--color-border);
    background: var(--color-surface);
    box-shadow:
      0 1px 0 color-mix(in oklch, var(--color-fg) 6%, transparent),
      0 16px 40px -16px rgb(0 0 0 / 0.55);
    padding: 0.375rem;
    transform-origin: top left;
  }
</style>
