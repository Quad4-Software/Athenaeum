<script lang="ts">
  import { Check, ChevronDown } from "@lucide/svelte";
  import { Select } from "bits-ui";

  export interface SelectOption {
    value: string;
    label: string;
    hint?: string;
    disabled?: boolean;
  }

  interface Props {
    value?: string;
    options: SelectOption[];
    label?: string;
    placeholder?: string;
    disabled?: boolean;
    minWidth?: number;
    class?: string;
    onchange?: (value: string) => void;
  }

  let {
    value = $bindable(""),
    options,
    label = "",
    placeholder = "Select…",
    disabled = false,
    minWidth = 220,
    class: className = "",
    onchange,
  }: Props = $props();

  let selected = $derived(options.find((o) => o.value === value));
  let display = $derived(selected?.label ?? placeholder);
</script>

<div class={className}>
  {#if label}
    <span class="select-label">{label}</span>
  {/if}
  <Select.Root
    type="single"
    bind:value
    {disabled}
    items={options.map((o) => ({ value: o.value, label: o.label, disabled: o.disabled }))}
    onValueChange={(v) => onchange?.(v)}
  >
    <Select.Trigger class="select-trigger" aria-label={label || placeholder}>
      <span class="select-value">{display}</span>
      <ChevronDown size={16} class="select-chevron" />
    </Select.Trigger>
    <Select.Portal>
      <Select.Content
        class="menu-panel select-panel"
        side="bottom"
        align="start"
        sideOffset={6}
        collisionPadding={8}
        style="min-width:{minWidth}px"
      >
        <Select.Viewport>
          {#each options as opt (opt.value)}
            <Select.Item
              value={opt.value}
              label={opt.label}
              disabled={opt.disabled}
              class="menu-item {opt.value === value ? 'menu-item--active' : ''}"
            >
              <span class="menu-item-label">{opt.label}</span>
              {#if opt.hint}
                <span class="menu-item-hint">{opt.hint}</span>
              {/if}
              {#if opt.value === value}
                <Check size={14} class="menu-item-check" />
              {/if}
            </Select.Item>
          {/each}
        </Select.Viewport>
      </Select.Content>
    </Select.Portal>
  </Select.Root>
</div>

<style>
  .select-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-muted);
  }

  :global(.select-trigger) {
    display: flex;
    width: 100%;
    min-width: 10rem;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    height: 2.25rem;
    border-radius: 0.5rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg);
    padding: 0 0.625rem;
    font-size: 0.875rem;
    color: var(--color-fg);
    cursor: pointer;
    transition:
      border-color 100ms ease,
      background-color 100ms ease;
  }

  :global(.select-trigger:hover:not(:disabled)) {
    background: var(--color-surface-hover);
  }

  :global(.select-trigger:disabled) {
    opacity: 0.55;
    cursor: default;
  }

  .select-value {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
  }

  .select-trigger :global(.select-chevron) {
    flex-shrink: 0;
    color: var(--color-muted);
  }

  :global(.select-panel) {
    animation: select-in 120ms ease-out;
  }

  @keyframes select-in {
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
