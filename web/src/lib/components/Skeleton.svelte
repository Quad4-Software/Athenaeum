<script lang="ts">
  interface Props {
    class?: string;
    width?: string;
    height?: string;
    rounded?: "sm" | "md" | "lg" | "full" | "card";
  }

  let { class: className = "", width = "100%", height = "1rem", rounded = "md" }: Props = $props();

  const radiusClass = $derived(
    {
      sm: "rounded-sm",
      md: "rounded-md",
      lg: "rounded-lg",
      full: "rounded-full",
      card: "rounded-[var(--radius-card)]",
    }[rounded],
  );
</script>

<div class="skeleton {radiusClass} {className}" style:width style:height aria-hidden="true"></div>

<style>
  .skeleton {
    background: linear-gradient(
      90deg,
      color-mix(in oklch, var(--color-border) 55%, transparent) 0%,
      color-mix(in oklch, var(--color-border-strong) 70%, transparent) 50%,
      color-mix(in oklch, var(--color-border) 55%, transparent) 100%
    );
    background-size: 200% 100%;
    animation: skeleton-shimmer 1.2s ease-in-out infinite;
  }

  @keyframes skeleton-shimmer {
    0% {
      background-position: 100% 0;
    }
    100% {
      background-position: -100% 0;
    }
  }
</style>
