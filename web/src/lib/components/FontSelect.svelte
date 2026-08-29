<script lang="ts">
  import { Check, ChevronDown } from "@lucide/svelte";
  import Popover from "./Popover.svelte";

  export interface FontOption {
    id: string;
    label: string;
    sample?: string;
    /** CSS font-family for preview. Falls back to inherit when omitted. */
    family?: string;
    disabled?: boolean;
  }

  interface Props {
    value?: string;
    options: FontOption[];
    label?: string;
    disabled?: boolean;
    minWidth?: number;
    class?: string;
    onchange?: (value: string) => void;
  }

  let {
    value = $bindable(""),
    options,
    label = "",
    disabled = false,
    minWidth = 280,
    class: className = "",
    onchange,
  }: Props = $props();

  let open = $state(false);

  let selected = $derived(options.find((o) => o.id === value) ?? options[0]);
</script>

<div class={className}>
  {#if label}
    <span class="font-select-label">{label}</span>
  {/if}
  <Popover bind:open {minWidth} align="start">
    {#snippet trigger(toggle)}
      <button
        type="button"
        class="font-select-trigger"
        {disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        onclick={toggle}
        style:font-family={selected?.family}
      >
        <span class="font-select-trigger-text">
          <span class="font-select-name">{selected?.label ?? "Select…"}</span>
          {#if selected?.sample}
            <span class="font-select-sample">{selected.sample}</span>
          {/if}
        </span>
        <ChevronDown size={16} class="font-select-chevron" />
      </button>
    {/snippet}

    <ul class="font-select-list" role="listbox" aria-label={label || "Font"}>
      {#each options as opt (opt.id)}
        <li role="none">
          <button
            type="button"
            role="option"
            class="font-select-option"
            class:font-select-option--active={opt.id === value}
            aria-selected={opt.id === value}
            disabled={opt.disabled}
            style:font-family={opt.family}
            onclick={() => {
              if (opt.disabled) return;
              value = opt.id;
              onchange?.(opt.id);
              open = false;
            }}
          >
            <span class="font-select-option-text">
              <span class="font-select-name">{opt.label}</span>
              {#if opt.sample}
                <span class="font-select-sample">{opt.sample}</span>
              {/if}
            </span>
            {#if opt.id === value}
              <Check size={14} class="font-select-check" />
            {/if}
          </button>
        </li>
      {/each}
    </ul>
  </Popover>
</div>

<style>
  .font-select-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-muted);
  }

  .font-select-trigger {
    display: flex;
    width: 100%;
    min-width: 12rem;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    min-height: 2.75rem;
    border-radius: 0.5rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg);
    padding: 0.5rem 0.625rem;
    color: var(--color-fg);
    cursor: pointer;
    transition:
      border-color 100ms ease,
      background-color 100ms ease;
  }

  .font-select-trigger:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .font-select-trigger:disabled {
    opacity: 0.55;
    cursor: default;
  }

  .font-select-trigger-text,
  .font-select-option-text {
    display: flex;
    min-width: 0;
    flex: 1;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
    text-align: left;
  }

  .font-select-name {
    font-size: 0.875rem;
    font-weight: 600;
    line-height: 1.25;
  }

  .font-select-sample {
    font-size: 0.8125rem;
    line-height: 1.3;
    color: var(--color-muted);
    font-weight: 400;
  }

  .font-select-trigger :global(.font-select-chevron) {
    flex-shrink: 0;
    color: var(--color-muted);
  }

  .font-select-list {
    margin: 0;
    padding: 0.25rem;
    list-style: none;
    max-height: min(22rem, 70vh);
    overflow-y: auto;
  }

  .font-select-option {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.5rem;
    border: 0;
    border-radius: 0.5rem;
    padding: 0.5rem 0.625rem;
    color: var(--color-fg);
    background: transparent;
    cursor: pointer;
    transition: background-color 100ms ease;
  }

  .font-select-option:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .font-select-option:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .font-select-option--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, transparent);
  }

  .font-select-option :global(.font-select-check) {
    flex-shrink: 0;
    color: var(--color-primary);
  }
</style>
