<script lang="ts">
  import { readerGestures } from "$lib/reader/reader-touch";

  interface Props {
    loading: boolean;
    onPrev: () => void;
    onNext: () => void;
    viewport?: HTMLDivElement;
    pagesRow?: HTMLDivElement;
  }

  let { loading, onPrev, onNext, viewport = $bindable(), pagesRow = $bindable() }: Props = $props();
</script>

<div
  bind:this={viewport}
  class="relative flex flex-1 justify-center overflow-auto bg-bg-elevated p-2 sm:p-4"
  style:touch-action="pan-x pan-y pinch-zoom"
  use:readerGestures={{
    onSwipeLeft: onNext,
    onSwipeRight: onPrev,
    onTapLeft: onPrev,
    onTapRight: onNext,
  }}
>
  {#if loading}
    <div class="absolute inset-0 z-10 grid place-items-center text-sm text-muted">
      Loading PDF...
    </div>
  {/if}
  <div
    bind:this={pagesRow}
    class="flex flex-wrap items-start justify-center gap-4"
    class:invisible={loading}
    aria-hidden={loading}
  ></div>
</div>
