<script lang="ts">
  import { Switch } from "bits-ui";

  interface Props {
    checked?: boolean;
    disabled?: boolean;
    label?: string;
    description?: string;
    id?: string;
    onchange?: (checked: boolean) => void;
  }

  let {
    checked = $bindable(false),
    disabled = false,
    label = "",
    description = "",
    id = "",
    onchange,
  }: Props = $props();
</script>

<label class="toggle-row" class:toggle-row--disabled={disabled} for={id || undefined}>
  <span class="toggle-copy">
    {#if label}
      <span class="toggle-label">{label}</span>
    {/if}
    {#if description}
      <span class="toggle-desc">{description}</span>
    {/if}
  </span>
  <Switch.Root
    {id}
    bind:checked
    {disabled}
    onCheckedChange={(v) => onchange?.(v)}
    class="toggle"
    aria-label={label || description || "Toggle"}
  >
    <Switch.Thumb class="toggle-thumb" />
  </Switch.Root>
</label>

<style>
  .toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    cursor: pointer;
  }

  .toggle-row--disabled {
    opacity: 0.55;
    cursor: default;
  }

  .toggle-copy {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 0.125rem;
  }

  .toggle-label {
    font-size: 0.875rem;
    color: var(--color-fg);
  }

  .toggle-desc {
    font-size: 0.75rem;
    color: var(--color-muted);
  }

  .toggle-row :global(.toggle) {
    position: relative;
    flex-shrink: 0;
    width: 2.5rem;
    height: 1.375rem;
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-bg-elevated);
    padding: 0;
    cursor: pointer;
    transition:
      background-color 120ms ease,
      border-color 120ms ease;
  }

  .toggle-row :global(.toggle:disabled) {
    cursor: default;
  }

  .toggle-row :global(.toggle[data-state="checked"]) {
    background: var(--color-primary);
    border-color: var(--color-primary);
  }

  .toggle-row :global(.toggle-thumb) {
    position: absolute;
    top: 2px;
    left: 2px;
    display: block;
    width: 1rem;
    height: 1rem;
    border-radius: 999px;
    background: var(--color-fg);
    transition: transform 120ms ease;
  }

  .toggle-row :global(.toggle[data-state="checked"] .toggle-thumb) {
    transform: translateX(1.125rem);
    background: var(--color-primary-fg);
  }
</style>
