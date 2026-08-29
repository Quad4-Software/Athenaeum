<script lang="ts">
  interface Props {
    url: string;
    initialPage?: number;
    onProgress?: (page: number, percent: number) => void;
  }

  let { url, initialPage = 1, onProgress }: Props = $props();

  let src = $derived(`${url}#page=${initialPage}&toolbar=1`);

  $effect(() => {
    onProgress?.(initialPage, 0);
  });
</script>

<iframe title="PDF reader" class="h-full w-full border-0 bg-bg-elevated" {src}></iframe>
