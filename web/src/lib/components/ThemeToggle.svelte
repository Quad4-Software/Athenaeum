<script lang="ts">
  import { Monitor, Moon, Sun } from "@lucide/svelte";
  import Popover from "./Popover.svelte";
  import MenuList from "./MenuList.svelte";
  import IconButton from "./IconButton.svelte";
  import { theme, type ThemePreference } from "$lib/stores/theme.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  let open = $state(false);

  const options: { id: ThemePreference; icon: typeof Sun }[] = [
    { id: "light", icon: Sun },
    { id: "dark", icon: Moon },
    { id: "system", icon: Monitor },
  ];

  let items = $derived(
    options.map((opt) => ({
      id: opt.id,
      label: i18n.t(`theme.${opt.id}`),
      icon: opt.icon,
      active: theme.preference === opt.id,
      onclick: () => {
        theme.set(opt.id);
        open = false;
      },
    })),
  );
</script>

<Popover bind:open align="end" minWidth={180}>
  {#snippet trigger(toggle)}
    <IconButton ariaLabel={i18n.t("theme.label")} title={i18n.t("theme.label")} onclick={toggle}>
      {#if theme.preference === "light"}
        <Sun size={18} />
      {:else if theme.preference === "dark"}
        <Moon size={18} />
      {:else}
        <Monitor size={18} />
      {/if}
    </IconButton>
  {/snippet}
  <MenuList title={i18n.t("theme.label")} {items} />
</Popover>
