<script lang="ts">
  import { ChevronLeft, ChevronRight } from "@lucide/svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { epubSpreadWidthClass, type EpubSpreadMode } from "$lib/reader/epub-reader";

  interface Props {
    spreadMode: EpubSpreadMode;
    surfaceBg: string;
    ready: boolean;
    loadError: string | null;
    onPrev: () => void;
    onNext: () => void;
    container?: HTMLDivElement;
  }

  let {
    spreadMode,
    surfaceBg,
    ready,
    loadError,
    onPrev,
    onNext,
    container = $bindable(),
  }: Props = $props();
</script>

<div class="epub-surface">
  <button
    type="button"
    class="epub-nav epub-nav--prev"
    aria-label="Previous page"
    onclick={onPrev}
  >
    <ChevronLeft size={24} />
  </button>

  <div class="epub-stage {epubSpreadWidthClass(spreadMode)}">
    <div
      bind:this={container}
      class="epub-mount"
      style:background-color={surfaceBg}
      style:touch-action="pan-y"
    ></div>
  </div>

  <button
    type="button"
    class="epub-nav epub-nav--next"
    aria-label="Next page"
    onclick={onNext}
  >
    <ChevronRight size={24} />
  </button>

  {#if loadError}
    <div class="epub-status epub-status--error">
      {loadError}
    </div>
  {:else if !ready}
    <div class="epub-status">
      {i18n.t("reader.loading")}
    </div>
  {/if}
</div>
