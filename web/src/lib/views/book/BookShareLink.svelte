<script lang="ts">
  import { X } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  interface Props {
    url: string;
    oncopy: () => void;
    ondismiss: () => void;
  }

  let { url, oncopy, ondismiss }: Props = $props();
</script>

<div
  class="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2"
>
  <input
    class="field-input min-w-0 flex-1 text-xs"
    type="text"
    readonly
    value={url}
    onclick={(e) => {
      const el = e.currentTarget;
      if (el instanceof HTMLInputElement) el.select();
    }}
  />
  <Button variant="ghost" size="sm" class="ring-1 ring-border" onclick={oncopy}>
    {i18n.t("book.copyLink")}
  </Button>
  <button
    type="button"
    class="btn btn-ghost text-xs"
    aria-label={i18n.t("book.dismissShare")}
    onclick={ondismiss}
  >
    <X size={14} />
  </button>
</div>
