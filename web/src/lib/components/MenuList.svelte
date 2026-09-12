<script lang="ts">
  import { Check } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import type { MenuItem } from "./menu";

  interface Props {
    title?: string;
    items?: MenuItem[];
    children?: Snippet;
  }

  let { title, items = [], children }: Props = $props();
</script>

<div class="menu">
  {#if title}
    <p class="menu-title" {title}>{title}</p>
  {/if}
  {#if items.length > 0}
    <ul class="menu-list" role="menu">
      {#each items as item (item.id)}
        <li role="none">
          {#if item.separator}
            <div class="menu-separator" role="separator"></div>
          {/if}
          <button
            type="button"
            role="menuitem"
            class="menu-item"
            class:menu-item--active={item.active}
            class:menu-item--danger={item.danger}
            disabled={item.disabled}
            onclick={item.onclick}
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
          </button>
        </li>
      {/each}
    </ul>
  {/if}
  {#if children}
    <div class="menu-body">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .menu-list {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .menu-item:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .menu-body {
    padding: 0.125rem;
  }
</style>
