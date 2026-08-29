<script lang="ts">
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

  function toggle() {
    if (disabled) return;
    checked = !checked;
    onchange?.(checked);
  }
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
  <button
    type="button"
    role="switch"
    {id}
    class="toggle"
    class:toggle--on={checked}
    aria-checked={checked}
    aria-label={label || description || "Toggle"}
    {disabled}
    onclick={toggle}
  >
    <span class="toggle-thumb"></span>
  </button>
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

  .toggle {
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

  .toggle:disabled {
    cursor: default;
  }

  .toggle--on {
    background: var(--color-primary);
    border-color: var(--color-primary);
  }

  .toggle-thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 1rem;
    height: 1rem;
    border-radius: 999px;
    background: var(--color-fg);
    transition: transform 120ms ease;
  }

  .toggle--on .toggle-thumb {
    transform: translateX(1.125rem);
    background: var(--color-primary-fg);
  }
</style>
