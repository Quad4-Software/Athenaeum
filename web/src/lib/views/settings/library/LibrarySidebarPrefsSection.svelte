<script lang="ts">
  import { ChevronDown, ChevronUp } from "@lucide/svelte";
  import { sidebarPrefs, SIDEBAR_SECTION_LABELS } from "$lib/stores/sidebar.svelte";
  import type { SidebarSectionId } from "$lib/api/types";
</script>

<div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
  <h2 class="text-sm font-semibold text-fg">Sidebar layout</h2>
  <p class="mt-1 text-sm text-muted">Show, hide, and reorder sidebar sections.</p>
  <ul class="mt-3 space-y-2">
    {#each sidebarPrefs.order as section (section)}
      <li class="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2">
        <label class="flex items-center gap-2 text-sm text-fg">
          <input
            type="checkbox"
            checked={!sidebarPrefs.isHidden(section)}
            onchange={() => sidebarPrefs.toggleSection(section)}
          />
          {SIDEBAR_SECTION_LABELS[section as SidebarSectionId]}
        </label>
        <div class="flex gap-1">
          <button
            type="button"
            class="btn btn-ghost text-xs ring-1 ring-border"
            aria-label="Move section up"
            onclick={() => sidebarPrefs.moveSection(section, -1)}
          >
            <ChevronUp size={14} />
          </button>
          <button
            type="button"
            class="btn btn-ghost text-xs ring-1 ring-border"
            aria-label="Move section down"
            onclick={() => sidebarPrefs.moveSection(section, 1)}
          >
            <ChevronDown size={14} />
          </button>
        </div>
      </li>
    {/each}
  </ul>
  <button
    type="button"
    class="btn btn-ghost mt-3 text-xs ring-1 ring-border"
    onclick={() => sidebarPrefs.reset()}
  >
    Reset sidebar
  </button>
</div>
