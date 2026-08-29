<script lang="ts">
  import type { Snippet } from "svelte";
  import { tick } from "svelte";

  interface Props {
    open?: boolean;
    placement?: "top" | "bottom";
    align?: "center" | "start" | "end";
    minWidth?: number;
    trigger: Snippet<[toggle: () => void]>;
    children: Snippet;
    onclose?: () => void;
  }

  let {
    open = $bindable(false),
    placement = "bottom",
    align = "center",
    minWidth = 0,
    trigger,
    children,
    onclose,
  }: Props = $props();

  let anchorEl = $state<HTMLDivElement>();
  let backdropEl = $state<HTMLButtonElement>();
  let panelEl = $state<HTMLDivElement>();
  let panelStyle = $state("");
  let panelReady = $state(false);

  function toggle() {
    open = !open;
  }

  function close() {
    if (!open) return;
    open = false;
    panelReady = false;
    onclose?.();
  }

  function portalNode(node: HTMLElement | undefined) {
    if (!node || typeof document === "undefined") return;
    if (node.parentElement !== document.body) {
      document.body.appendChild(node);
    }
  }

  function reposition() {
    if (!open || !anchorEl || !panelEl) return;

    const rect = anchorEl.getBoundingClientRect();
    const panel = panelEl.getBoundingClientRect();
    const margin = 8;
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const panelWidth = Math.max(panel.width, minWidth, 1);
    const panelHeight = Math.max(panel.height, 1);

    let left = rect.left;
    if (align === "center") {
      left = rect.left + rect.width / 2 - panelWidth / 2;
    } else if (align === "end") {
      left = rect.right - panelWidth;
    }

    left = Math.max(margin, Math.min(left, vw - panelWidth - margin));

    let top = placement === "top" ? rect.top - panelHeight - margin : rect.bottom + margin;
    if (placement === "top" && top < margin) {
      top = rect.bottom + margin;
    } else if (placement === "bottom" && top + panelHeight > vh - margin) {
      top = rect.top - panelHeight - margin;
    }
    top = Math.max(margin, Math.min(top, vh - panelHeight - margin));

    const widthRule = minWidth > 0 ? `min-width:${minWidth}px;` : "";
    panelStyle = `top:${top}px;left:${left}px;${widthRule}`;
  }

  async function layoutPanel() {
    panelReady = false;
    await tick();
    portalNode(backdropEl);
    portalNode(panelEl);
    reposition();
    await tick();
    reposition();
    requestAnimationFrame(() => {
      reposition();
      panelReady = true;
    });
  }

  $effect(() => {
    if (!open) return;

    let resizeObserver: ResizeObserver | undefined;

    void (async () => {
      await layoutPanel();
      if (panelEl) {
        resizeObserver = new ResizeObserver(() => reposition());
        resizeObserver.observe(panelEl);
      }
    })();

    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close();
    };
    const onScroll = () => reposition();
    const onResize = () => reposition();

    window.addEventListener("keydown", onKey);
    window.addEventListener("resize", onResize);
    window.addEventListener("scroll", onScroll, true);

    return () => {
      resizeObserver?.disconnect();
      window.removeEventListener("keydown", onKey);
      window.removeEventListener("resize", onResize);
      window.removeEventListener("scroll", onScroll, true);
    };
  });
</script>

{#if open}
  <button
    type="button"
    class="popover-backdrop"
    aria-label="Close menu"
    bind:this={backdropEl}
    onclick={close}
  ></button>
{/if}

<div class="popover-anchor" class:popover-anchor--open={open} bind:this={anchorEl}>
  {@render trigger(toggle)}
</div>

{#if open}
  <div
    class="popover-panel"
    class:popover-panel--top={placement === "top"}
    class:popover-panel--ready={panelReady}
    bind:this={panelEl}
    style={panelStyle}
    role="dialog"
    aria-modal="true"
  >
    {@render children()}
  </div>
{/if}

<style>
  .popover-backdrop {
    position: fixed;
    inset: 0;
    z-index: 40;
    border: 0;
    padding: 0;
    background: transparent;
    cursor: default;
  }

  .popover-anchor {
    display: inline-flex;
    position: relative;
  }

  .popover-anchor--open {
    z-index: 45;
  }

  .popover-panel {
    position: fixed;
    z-index: 50;
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
    opacity: 0;
    pointer-events: none;
  }

  .popover-panel--ready {
    opacity: 1;
    pointer-events: auto;
  }

  .popover-panel--top {
    transform-origin: bottom center;
    animation: popover-in-up 140ms ease-out;
  }

  .popover-panel:not(.popover-panel--top) {
    transform-origin: top center;
    animation: popover-in-down 140ms ease-out;
  }

  @keyframes popover-in-up {
    from {
      opacity: 0;
      transform: translateY(6px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  @keyframes popover-in-down {
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
