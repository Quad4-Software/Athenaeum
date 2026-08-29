<script lang="ts">
  import { ChevronDown } from "@lucide/svelte";
  import Popover from "./Popover.svelte";
  import MenuList, { type MenuItem } from "./MenuList.svelte";

  export interface SelectOption {
    value: string;
    label: string;
    hint?: string;
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

  let open = $state(false);

  let selected = $derived(options.find((o) => o.value === value));
  let display = $derived(selected?.label ?? placeholder);

  let items = $derived(
    options.map((opt): MenuItem => ({
      id: opt.value,
      label: opt.label,
      hint: opt.hint,
      active: opt.value === value,
      onclick: () => {
        value = opt.value;
        onchange?.(opt.value);
        open = false;
      },
    })),
  );
</script>

<div class={className}>
  {#if label}
    <span class="select-label">{label}</span>
  {/if}
  <Popover bind:open {minWidth} align="start">
    {#snippet trigger(toggle)}
      <button
        type="button"
        class="select-trigger"
        {disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        onclick={toggle}
      >
        <span class="select-value">{display}</span>
        <ChevronDown size={16} class="select-chevron" />
      </button>
    {/snippet}
    <MenuList title={label || undefined} {items} />
  </Popover>
</div>

<style>
  .select-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-muted);
  }

  .select-trigger {
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

  .select-trigger:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .select-trigger:disabled {
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
</style>
