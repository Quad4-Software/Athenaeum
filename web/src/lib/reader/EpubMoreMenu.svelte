<script lang="ts">
  import { Bookmark, Highlighter, List, MoreVertical, Volume2 } from "@lucide/svelte";
  import Dropdown from "$lib/components/Dropdown.svelte";
  import type { MenuItem } from "$lib/components/menu";
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

  let menuItems = $derived<MenuItem[]>([
    ...(bookId
      ? [
          {
            id: "narrate",
            label: narrator.active ? i18n.t("narrator.stop") : i18n.t("narrator.play"),
            icon: Volume2,
            onclick: () => void onNarrate(),
          },
          {
            id: "bookmark",
            label: i18n.t("reader.bookmarkLabel"),
            icon: Bookmark,
            onclick: onBookmark,
          },
          {
            id: "highlight",
            label: "Highlight",
            icon: Highlighter,
            disabled: !selectionCfi,
            onclick: onHighlight,
          },
          {
            id: "annotations",
            label: "Annotations",
            icon: List,
            onclick: onToggleAnnotations,
          },
        ]
      : []),
    {
      id: "shortcuts",
      label: "Keyboard shortcuts",
      onclick: onOpenShortcuts,
    },
  ]);
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
    <button class="btn btn-ghost text-xs" onclick={onToggleAnnotations} aria-label="Annotations">
      <List size={16} />
    </button>
  {/if}
  <button class="btn btn-ghost text-xs" aria-label="Keyboard shortcuts" onclick={onOpenShortcuts}>
    ?
  </button>
</div>

<Dropdown bind:open side="bottom" align="end" minWidth={200} items={menuItems}>
  {#snippet trigger(props)}
    <button
      type="button"
      class="btn btn-ghost md:hidden"
      class:ring-1={open}
      class:ring-border={open}
      aria-label="More options"
      {...props}
    >
      <MoreVertical size={16} />
    </button>
  {/snippet}
</Dropdown>
