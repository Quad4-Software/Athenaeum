<script lang="ts">
  import Skeleton from "./Skeleton.svelte";
</script>

<div class="flex flex-col gap-2">
  <Skeleton rounded="card" height="0" class="aspect-[2/3] w-full" />
  <Skeleton height="0.875rem" width="88%" />
  <Skeleton height="0.75rem" width="62%" />
</div>
