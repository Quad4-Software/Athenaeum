<script lang="ts">
  import { Monitor, Moon, Sun } from "@lucide/svelte";
  import Dropdown from "./Dropdown.svelte";
  import IconButton from "./IconButton.svelte";
  import type { MenuItem } from "./menu";
  import { theme, type ThemePreference } from "$lib/stores/theme.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  const options: { id: ThemePreference; icon: typeof Sun }[] = [
    { id: "light", icon: Sun },
    { id: "dark", icon: Moon },
    { id: "system", icon: Monitor },
  ];

  let items = $derived<MenuItem[]>(
    options.map((opt) => ({
      id: opt.id,
      label: i18n.t(`theme.${opt.id}`),
      icon: opt.icon,
      active: theme.preference === opt.id,
      onclick: () => theme.set(opt.id),
    })),
  );
</script>

<Dropdown align="end" minWidth={180} title={i18n.t("theme.label")} {items}>
  {#snippet trigger(props)}
    <IconButton {...props} ariaLabel={i18n.t("theme.label")} title={i18n.t("theme.label")}>
      {#if theme.preference === "light"}
        <Sun size={18} />
      {:else if theme.preference === "dark"}
        <Moon size={18} />
      {:else}
        <Monitor size={18} />
      {/if}
    </IconButton>
  {/snippet}
</Dropdown>
