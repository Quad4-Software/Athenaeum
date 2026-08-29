import { router } from "$lib/router.svelte";
import { commandPalette } from "$lib/stores/commandPalette.svelte";
import { keybindings } from "$lib/stores/keybindings.svelte";
import { isTypingTarget } from "./chords";
import { runCommand } from "./registry";

/**
 * Global shell keydown handler. Mount once from App when the main shell is active.
 * Ignores typing targets except for palette.open (Mod+K).
 */
export function handleShellKeydown(event: KeyboardEvent): void {
  if (event.defaultPrevented) return;
  if (router.current.name === "reader") return;

  const matched = keybindings.matchEvent(event);
  if (!matched) return;

  if (matched === "palette.open") {
    event.preventDefault();
    void runCommand(matched);
    return;
  }

  if (commandPalette.open) return;
  if (isTypingTarget(event.target)) return;

  event.preventDefault();
  void runCommand(matched);
}
