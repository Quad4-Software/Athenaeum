<script lang="ts">
  import { Check } from "@lucide/svelte";
  import { DropdownMenu } from "bits-ui";
  import type { MenuItem } from "./menu";

  interface Props {
    title?: string;
    items?: MenuItem[];
  }

  let { title, items = [] }: Props = $props();
</script>

{#if title}
  <DropdownMenu.GroupHeading class="menu-title" {title}>{title}</DropdownMenu.GroupHeading>
{/if}
{#each items as item (item.id)}
  {#if item.separator}
    <DropdownMenu.Separator class="menu-separator" />
  {/if}
  <DropdownMenu.Item
    class="menu-item {item.active ? 'menu-item--active' : ''} {item.danger
      ? 'menu-item--danger'
      : ''}"
    disabled={item.disabled}
    textValue={item.label}
    onSelect={() => item.onclick?.()}
  >
    <span class="menu-item-icon-slot" aria-hidden="true">
      {#if item.icon}
        {@const ItemIcon = item.icon}
        <ItemIcon size={16} class="menu-item-icon" />
      {/if}
    </span>
    <span class="menu-item-label">{item.label}</span>
    {#if item.hint}
      <span class="menu-item-hint">{item.hint}</span>
    {/if}
    {#if item.active}
      <Check size={14} class="menu-item-check" />
    {/if}
  </DropdownMenu.Item>
{/each}
