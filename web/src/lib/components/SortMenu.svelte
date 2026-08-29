<script lang="ts">
  import { ArrowDownWideNarrow } from "@lucide/svelte";
  import { library } from "$lib/stores/library.svelte";
  import type { SortKey } from "$lib/api/types";
  import { i18n } from "$lib/stores/i18n.svelte";

  const sortKeys: SortKey[] = ["recent", "oldest", "title", "author", "progress"];

  let options = $derived(
    sortKeys
      .filter((value) => value !== "progress" || library.inProgressFilter)
      .map((value) => ({
        value,
        label: i18n.t(`sort.${value}`),
      })),
  );
</script>

<label class="relative flex shrink-0 items-center" title={i18n.t("sort.label")}>
  <span class="pointer-events-none absolute left-2 text-subtle sm:left-2.5">
    <ArrowDownWideNarrow size={16} />
  </span>
  <select
    class="h-11 w-11 appearance-none rounded-lg border border-border bg-surface pl-8 pr-1 text-sm text-transparent focus:border-primary focus:outline-none sm:h-10 sm:w-auto sm:pr-8 sm:text-fg"
    value={library.sort}
    onchange={(e) => library.setSort(e.currentTarget.value as SortKey)}
    aria-label={i18n.t("sort.label")}
  >
    {#each options as opt (opt.value)}
      <option value={opt.value}>{opt.label}</option>
    {/each}
  </select>
</label>
