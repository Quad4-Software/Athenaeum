<script lang="ts">
  import KeybindingsSettings from "$lib/components/KeybindingsSettings.svelte";
  import LibraryAppearanceSection from "./library/LibraryAppearanceSection.svelte";
  import LibraryMountsSection from "./library/LibraryMountsSection.svelte";
  import LibrarySidebarPrefsSection from "./library/LibrarySidebarPrefsSection.svelte";
  import LibraryStatsSection from "./library/LibraryStatsSection.svelte";
</script>

<div class="space-y-6">
  <LibraryAppearanceSection />
  <LibraryMountsSection />
  <LibrarySidebarPrefsSection />
  <LibraryStatsSection />
  <KeybindingsSettings />
</div>
