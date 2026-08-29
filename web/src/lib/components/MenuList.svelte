<script lang="ts">
  import { Check } from "@lucide/svelte";
  import type { Component, Snippet } from "svelte";

  export interface MenuItem {
    id: string;
    label: string;
    hint?: string;
    active?: boolean;
    danger?: boolean;
    disabled?: boolean;
    separator?: boolean;
    icon?: Component<{ size?: number; class?: string }>;
    onclick?: () => void;
  }

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
  .menu-title {
    margin: 0;
    padding: 0.2rem 0.5rem 0.3rem;
    font-size: 0.6875rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-subtle);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .menu-list {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .menu-separator {
    height: 1px;
    margin: 0.25rem 0.375rem;
    background: var(--color-border);
  }

  .menu-item {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.4rem;
    border: 0;
    border-radius: 0.5rem;
    padding: 0.4375rem 0.5rem;
    font-size: 0.8125rem;
    text-align: left;
    color: var(--color-fg);
    background: transparent;
    cursor: pointer;
    transition: background-color 100ms ease;
  }

  .menu-item:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .menu-item:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .menu-item--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, transparent);
  }

  .menu-item--danger {
    color: var(--color-danger);
  }

  .menu-item-icon-slot {
    display: grid;
    flex-shrink: 0;
    place-items: center;
    width: 1rem;
    height: 1rem;
  }

  .menu-item-label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .menu-item :global(.menu-item-icon) {
    flex-shrink: 0;
    color: var(--color-muted);
  }

  .menu-item--active :global(.menu-item-icon) {
    color: var(--color-primary);
  }

  .menu-item--danger :global(.menu-item-icon) {
    color: inherit;
  }

  .menu-item-hint {
    flex-shrink: 0;
    max-width: 4.5rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.6875rem;
    color: var(--color-muted);
    font-variant-numeric: tabular-nums;
  }

  .menu-item :global(.menu-item-check) {
    color: var(--color-primary);
    flex-shrink: 0;
  }

  .menu-body {
    padding: 0.125rem;
  }
</style>
