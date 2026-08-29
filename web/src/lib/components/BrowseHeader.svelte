<script lang="ts">
  interface Props {
    title: string;
    subtitle?: string;
    countLabel?: string;
    progressLabel?: string;
  }

  let { title, subtitle = "", countLabel = "", progressLabel = "" }: Props = $props();
</script>

<header class="mb-5 flex flex-wrap items-end justify-between gap-3">
  <div class="min-w-0">
    <h1 class="font-display text-2xl font-semibold tracking-tight text-fg sm:text-3xl">{title}</h1>
    {#if subtitle}
      <p class="mt-1 text-sm text-muted">{subtitle}</p>
    {/if}
  </div>
  {#if countLabel || progressLabel}
    <p class="text-xs text-subtle sm:text-sm">
      {#if countLabel}<span>{countLabel}</span>{/if}
      {#if countLabel && progressLabel}<span class="mx-1.5 text-border-strong">·</span>{/if}
      {#if progressLabel}<span>{progressLabel}</span>{/if}
    </p>
  {/if}
</header>
