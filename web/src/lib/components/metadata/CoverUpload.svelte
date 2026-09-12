<script lang="ts">
  import { Trash2, Upload } from "@lucide/svelte";
  import Cover from "$lib/components/Cover.svelte";
  import { api, ApiError } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Book } from "$lib/api/types";

  interface Props {
    book: Book;
    /** Bump to force the cover image to reload. */
    coverKey: number;
    onuploaded?: (book: Book) => void;
    onremoved?: (book: Book) => void;
  }

  let { book, coverKey, onuploaded, onremoved }: Props = $props();

  let uploading = $state(false);

  async function onCoverSelected(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    uploading = true;
    try {
      const updated = await api.uploadCover(book.id, file);
      onuploaded?.(updated);
      toast.success(i18n.t("book.coverUpdated"));
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.coverUploadFailed"));
    } finally {
      uploading = false;
    }
  }

  async function removeCover() {
    uploading = true;
    try {
      const updated = await api.deleteCover(book.id);
      onremoved?.(updated);
      toast.success(i18n.t("book.coverRemoved"));
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.coverRemoveFailed"));
    } finally {
      uploading = false;
    }
  }
</script>

<div class="w-32 shrink-0">
  {#key coverKey}
    <Cover book={{ ...book, modifiedAt: book.modifiedAt }} />
  {/key}
  <div class="mt-2 flex flex-col gap-2">
    <label class="btn btn-ghost ring-1 ring-border cursor-pointer text-xs">
      <Upload size={14} />
      {uploading ? "Uploading..." : "Upload cover"}
      <input
        type="file"
        accept="image/jpeg,image/png,image/webp"
        class="sr-only"
        disabled={uploading}
        onchange={onCoverSelected}
      />
    </label>
    {#if book.hasCover}
      <button
        type="button"
        class="btn btn-ghost text-xs text-danger"
        disabled={uploading}
        onclick={removeCover}
      >
        <Trash2 size={14} />
        Remove cover
      </button>
    {/if}
  </div>
</div>
