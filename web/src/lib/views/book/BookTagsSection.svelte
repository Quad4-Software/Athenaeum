<script lang="ts">
  import { Plus, Tag, X } from "@lucide/svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  interface Props {
    tags: string[];
    tagInput: string;
    addingTag: boolean;
    onfilter: (name: string) => void;
    onremove: (name: string) => void;
    onadd: () => void;
  }

  let { tags, tagInput = $bindable(), addingTag, onfilter, onremove, onadd }: Props = $props();
</script>

<div class="mt-4">
  <p class="mb-1.5 flex items-center gap-1.5 text-sm text-muted">
    <Tag size={14} />
    {i18n.t("book.tags")}
  </p>
  <div class="flex flex-wrap items-center gap-2">
    {#each tags as tagName (tagName)}
      <span
        class="inline-flex items-center gap-1 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-fg"
      >
        <button type="button" class="hover:underline" onclick={() => onfilter(tagName)}>
          {tagName}
        </button>
        <button
          type="button"
          class="text-subtle hover:text-danger"
          aria-label={i18n.t("book.removeTag", { name: tagName })}
          onclick={() => onremove(tagName)}
        >
          <X size={12} />
        </button>
      </span>
    {/each}
    <form
      class="inline-flex items-center gap-1"
      onsubmit={(e) => {
        e.preventDefault();
        onadd();
      }}
    >
      <input
        type="text"
        class="field-input h-8 w-28 text-xs"
        placeholder={i18n.t("book.addTagPlaceholder")}
        bind:value={tagInput}
        disabled={addingTag}
      />
      <button
        type="submit"
        class="btn btn-ghost min-h-8 px-2 text-xs ring-1 ring-border"
        disabled={addingTag || !tagInput.trim()}
      >
        <Plus size={14} />
      </button>
    </form>
  </div>
</div>
