<script lang="ts">
  import { Check, ChevronDown } from "@lucide/svelte";
  import { Select } from "bits-ui";

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

  let selected = $derived(options.find((o) => o.id === value) ?? options[0]);
</script>

<div class={className}>
  {#if label}
    <span class="font-select-label">{label}</span>
  {/if}
  <Select.Root
    type="single"
    bind:value
    {disabled}
    items={options.map((o) => ({ value: o.id, label: o.label, disabled: o.disabled }))}
    onValueChange={(v) => onchange?.(v)}
  >
    <Select.Trigger
      class="font-select-trigger"
      aria-label={label || "Font"}
      style={selected?.family ? `font-family:${selected.family}` : undefined}
    >
      <span class="font-select-trigger-text">
        <span class="font-select-name">{selected?.label ?? "Select…"}</span>
        {#if selected?.sample}
          <span class="font-select-sample">{selected.sample}</span>
        {/if}
      </span>
      <ChevronDown size={16} class="font-select-chevron" />
    </Select.Trigger>
    <Select.Portal>
      <Select.Content
        class="menu-panel font-select-panel"
        side="bottom"
        align="start"
        sideOffset={6}
        collisionPadding={8}
        style="min-width:{minWidth}px"
      >
        <Select.Viewport>
          {#each options as opt (opt.id)}
            <Select.Item
              value={opt.id}
              label={opt.label}
              disabled={opt.disabled}
              class="font-select-option {opt.id === value ? 'font-select-option--active' : ''}"
              style={opt.family ? `font-family:${opt.family}` : undefined}
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
            </Select.Item>
          {/each}
        </Select.Viewport>
      </Select.Content>
    </Select.Portal>
  </Select.Root>
</div>

<style>
  .font-select-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-muted);
  }

  :global(.font-select-trigger) {
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

  :global(.font-select-trigger:hover:not(:disabled)) {
    background: var(--color-surface-hover);
  }

  :global(.font-select-trigger:disabled) {
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

  :global(.font-select-panel) {
    max-height: min(22rem, 70vh);
    animation: font-select-in 120ms ease-out;
  }

  :global(.font-select-option) {
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
    outline: none;
  }

  :global(.font-select-option[data-highlighted]) {
    background: var(--color-surface-hover);
  }

  :global(.font-select-option[data-disabled]) {
    opacity: 0.45;
    cursor: default;
  }

  :global(.font-select-option--active) {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, transparent);
  }

  :global(.font-select-option .font-select-check) {
    flex-shrink: 0;
    color: var(--color-primary);
  }

  @keyframes font-select-in {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
