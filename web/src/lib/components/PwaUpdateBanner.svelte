<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { pwa } from "$lib/stores/pwa.svelte";
</script>

{#if pwa.updateAvailable}
  <div
    class="border-b border-accent/40 bg-accent/10 px-4 py-2 text-center text-sm text-fg"
    role="status"
  >
    <span>{i18n.t("app.updateAvailable")}</span>
    <span class="mx-2 inline-flex gap-2">
      <Button size="sm" onclick={() => pwa.applyUpdate()}>
        {i18n.t("app.updateReload")}
      </Button>
      <Button size="sm" variant="ghost" onclick={() => pwa.dismissUpdate()}>
        {i18n.t("app.updateLater")}
      </Button>
    </span>
  </div>
{/if}
