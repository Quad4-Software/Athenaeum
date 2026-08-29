<script lang="ts">
  import { sanitizeHtml } from "$lib/utils/sanitize-html";

  interface Props {
    value: string;
  }

  let { value }: Props = $props();
</script>

<!-- output is sanitized in sanitizeHtml() -->
<div class="book-description mt-6 text-sm leading-relaxed text-muted">
  {@html sanitizeHtml(value)}
</div>

<style>
  :global(.book-description p) {
    margin-top: 0.75rem;
  }
  :global(.book-description p:first-child) {
    margin-top: 0;
  }
  :global(.book-description h3) {
    margin-top: 1rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-fg);
  }
</style>
