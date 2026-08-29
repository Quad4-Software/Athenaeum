<script lang="ts">
  import { brand } from "$lib/brand";
  import { connectivity } from "$lib/stores/connectivity.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  let message = $derived(
    !connectivity.online
      ? i18n.t("connectivity.offline")
      : i18n.t("connectivity.unreachable", { app: brand.appName }),
  );
</script>

{#if connectivity.disconnected}
  <div
    class="border-b border-danger/40 bg-danger/10 px-4 py-2 text-center text-sm text-fg"
    role="status"
  >
    {message}
  </div>
{/if}
