<script lang="ts">
  import { i18n } from "$lib/stores/i18n.svelte";
  import { keybindings } from "$lib/stores/keybindings.svelte";
  import { listCommands } from "$lib/commands/registry";
  import { chordFromEvent, formatChord, isMacPlatform, isTypingTarget } from "$lib/commands/chords";
  import type { CommandId } from "$lib/commands/types";
  import { toast } from "$lib/stores/toast.svelte";

  const isMac = isMacPlatform();

  let recordingId = $state<CommandId | null>(null);

  let shellCommands = $derived(listCommands().filter((c) => c.scope === "shell"));

  function labelFor(id: CommandId): string {
    return i18n.t(listCommands().find((c) => c.id === id)?.titleKey ?? id);
  }

  function startRecord(id: CommandId) {
    recordingId = id;
  }

  function cancelRecord() {
    recordingId = null;
  }

  function onWindowKeydown(event: KeyboardEvent) {
    if (!recordingId) return;
    if (isTypingTarget(event.target) && event.key !== "Escape") return;
    event.preventDefault();
    event.stopPropagation();

    if (event.key === "Escape") {
      cancelRecord();
      return;
    }
    if (event.key === "Backspace" || event.key === "Delete") {
      keybindings.setBinding(recordingId, null);
      recordingId = null;
      return;
    }

    const chord = chordFromEvent(event);
    if (
      chord.key === "Shift" ||
      chord.key === "Control" ||
      chord.key === "Meta" ||
      chord.key === "Alt"
    ) {
      return;
    }

    const conflict = keybindings.conflictFor(chord, recordingId);
    if (conflict) {
      keybindings.setBinding(conflict, null);
      toast.info(i18n.t("commands.rebindConflict", { command: labelFor(conflict) }));
    }
    keybindings.setBinding(recordingId, chord);
    recordingId = null;
  }

  function bindingLabel(id: CommandId): string {
    const chord = keybindings.bindingFor(id);
    if (!chord) return i18n.t("commands.unbound");
    return formatChord(chord, isMac);
  }

  function resetOne(id: CommandId) {
    keybindings.reset(id);
  }

  function resetAll() {
    keybindings.resetAll();
    toast.success(i18n.t("commands.resetAllDone"));
  }
</script>

<svelte:window onkeydown={onWindowKeydown} />

<section id="settings-keyboard-shortcuts" class="mt-10 space-y-4">
  <div class="flex flex-wrap items-end justify-between gap-3">
    <div>
      <h2 class="font-display text-xl font-semibold text-fg">{i18n.t("commands.settingsTitle")}</h2>
      <p class="mt-1 text-sm text-muted">{i18n.t("commands.settingsBody")}</p>
    </div>
    <button type="button" class="btn btn-ghost text-xs ring-1 ring-border" onclick={resetAll}>
      {i18n.t("commands.resetAll")}
    </button>
  </div>

  <ul
    class="divide-y divide-border overflow-hidden rounded-[var(--radius-card)] border border-border bg-bg-elevated/40"
  >
    {#each shellCommands as cmd (cmd.id)}
      <li class="flex flex-wrap items-center gap-3 px-3 py-2.5 sm:px-4">
        <div class="min-w-0 flex-1">
          <p class="text-sm text-fg">{i18n.t(cmd.titleKey)}</p>
          {#if keybindings.isCustom(cmd.id)}
            <p class="text-[0.7rem] text-subtle">{i18n.t("commands.customized")}</p>
          {/if}
        </div>
        <kbd
          class="rounded border border-border bg-bg px-2 py-1 font-mono text-xs text-muted"
          class:ring-1={recordingId === cmd.id}
          class:ring-primary={recordingId === cmd.id}
        >
          {recordingId === cmd.id ? i18n.t("commands.pressKeys") : bindingLabel(cmd.id)}
        </kbd>
        <div class="flex gap-1">
          <button
            type="button"
            class="btn btn-ghost text-xs"
            onclick={() => (recordingId === cmd.id ? cancelRecord() : startRecord(cmd.id))}
          >
            {recordingId === cmd.id ? i18n.t("confirm.cancel") : i18n.t("commands.rebind")}
          </button>
          {#if keybindings.isCustom(cmd.id)}
            <button type="button" class="btn btn-ghost text-xs" onclick={() => resetOne(cmd.id)}>
              {i18n.t("commands.resetOne")}
            </button>
          {/if}
        </div>
      </li>
    {/each}
  </ul>
  {#if recordingId}
    <p class="text-xs text-muted">{i18n.t("commands.recordHint")}</p>
  {/if}
</section>
