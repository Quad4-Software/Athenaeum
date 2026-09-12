<script lang="ts">
  import { Slider } from "bits-ui";

  interface Props {
    value?: number;
    min?: number;
    max?: number;
    step?: number;
    disabled?: boolean;
    ariaLabel: string;
    class?: string;
    onchange?: (value: number) => void;
    oncommit?: (value: number) => void;
  }

  let {
    value = $bindable(0),
    min = 0,
    max = 100,
    step = 1,
    disabled = false,
    ariaLabel,
    class: className = "",
    onchange,
    oncommit,
  }: Props = $props();
</script>

<Slider.Root
  type="single"
  bind:value
  {min}
  {max}
  {step}
  {disabled}
  class="slider {className}"
  onValueChange={(v) => onchange?.(v)}
  onValueCommit={(v) => oncommit?.(v)}
>
  <Slider.Range class="slider-range" />
  <Slider.Thumb index={0} class="slider-thumb" aria-label={ariaLabel} />
</Slider.Root>

<style>
  :global(.slider) {
    position: relative;
    display: flex;
    width: 100%;
    height: 1.25rem;
    align-items: center;
    touch-action: none;
    user-select: none;
    cursor: pointer;
  }

  :global(.slider[data-disabled]) {
    opacity: 0.5;
    cursor: default;
  }

  :global(.slider::before) {
    content: "";
    position: absolute;
    left: 0;
    right: 0;
    height: 0.375rem;
    border-radius: 999px;
    background: var(--color-border);
  }

  :global(.slider .slider-range) {
    position: absolute;
    height: 0.375rem;
    border-radius: 999px;
    background: var(--color-primary);
  }

  :global(.slider .slider-thumb) {
    display: block;
    width: 1rem;
    height: 1rem;
    border-radius: 999px;
    border: 2px solid var(--color-primary);
    background: var(--color-bg-elevated);
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.2);
    cursor: grab;
  }

  :global(.slider .slider-thumb:focus-visible) {
    outline: 2px solid var(--ring);
    outline-offset: 2px;
  }
</style>
