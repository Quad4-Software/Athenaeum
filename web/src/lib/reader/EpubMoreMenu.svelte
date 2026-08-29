<script lang="ts">
  import { Bookmark, Highlighter, List, MoreVertical, Volume2 } from "@lucide/svelte";
  import Popover from "$lib/components/Popover.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";

  interface Props {
    open: boolean;
    bookId?: number;
    selectionCfi: string;
    onNarrate: () => void;
    onBookmark: () => void;
    onHighlight: () => void;
    onToggleAnnotations: () => void;
    onOpenShortcuts: () => void;
  }

  let {
    open = $bindable(false),
    bookId,
    selectionCfi,
    onNarrate,
    onBookmark,
    onHighlight,
    onToggleAnnotations,
    onOpenShortcuts,
  }: Props = $props();
</script>

<div class="hidden items-center gap-1 md:flex">
  {#if bookId}
    <button
      class="btn btn-ghost text-xs"
      onclick={() => void onNarrate()}
      aria-label={narrator.active ? i18n.t("narrator.stop") : i18n.t("narrator.play")}
      aria-pressed={narrator.active}
    >
      <Volume2 size={16} />
    </button>
    <button class="btn btn-ghost text-xs" onclick={onBookmark} aria-label="Bookmark">
      <Bookmark size={16} />
    </button>
    <button
      class="btn btn-ghost text-xs"
      onclick={onHighlight}
      aria-label="Highlight"
      disabled={!selectionCfi}
    >
      <Highlighter size={16} />
    </button>
    <button
      class="btn btn-ghost text-xs"
      onclick={onToggleAnnotations}
      aria-label="Annotations"
    >
      <List size={16} />
    </button>
  {/if}
  <button
    class="btn btn-ghost text-xs"
    aria-label="Keyboard shortcuts"
    onclick={onOpenShortcuts}
  >
    ?
  </button>
</div>

<Popover bind:open placement="bottom" align="end" minWidth={200}>
  {#snippet trigger(toggle)}
    <button
      type="button"
      class="btn btn-ghost md:hidden"
      class:ring-1={open}
      class:ring-border={open}
      aria-expanded={open}
      aria-label="More options"
      onclick={toggle}
    >
      <MoreVertical size={16} />
    </button>
  {/snippet}
  <div class="flex flex-col gap-1 p-1 md:hidden">
    {#if bookId}
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        onclick={() => {
          open = false;
          void onNarrate();
        }}
      >
        <Volume2 size={14} />
        {narrator.active ? i18n.t("narrator.stop") : i18n.t("narrator.play")}
      </button>
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        onclick={() => {
          open = false;
          void onBookmark();
        }}
      >
        <Bookmark size={14} />
        {i18n.t("reader.bookmarkLabel")}
      </button>
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        onclick={() => {
          open = false;
          void onHighlight();
        }}
        disabled={!selectionCfi}
      >
        <Highlighter size={14} /> Highlight
      </button>
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        onclick={() => {
          open = false;
          onToggleAnnotations();
        }}
      >
        <List size={14} /> Annotations
      </button>
    {/if}
    <button
      class="btn btn-ghost w-full justify-start text-xs"
      onclick={() => {
        open = false;
        onOpenShortcuts();
      }}
    >
      Keyboard shortcuts
    </button>
  </div>
</Popover>
