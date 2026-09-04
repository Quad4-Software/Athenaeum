<script lang="ts">
  import { CloudOff } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { BookOfflineStatus } from "$lib/offline/book-cache";

  interface Props {
    busy: boolean;
    status: BookOfflineStatus;
    onclick: () => void;
  }

  let { busy, status, onclick }: Props = $props();
</script>

<Button
  variant="ghost"
  class="min-h-11 ring-1 ring-border"
  loading={busy || status.downloading}
  {onclick}
>
  <CloudOff size={16} />
  {#if status.complete}
    {i18n.t("book.offlineReady")}
  {:else if status.downloading}
    {i18n.t("book.offlineCaching", {
      pct: String(Math.round((status.cachedBytes / Math.max(status.totalBytes, 1)) * 100)),
    })}
  {:else}
    {i18n.t("book.saveOffline")}
  {/if}
</Button>
