/**
 * Persistable keybinding overrides for shell commands.
 */

import { storageKey } from "$lib/brand/storage";
import type { CommandId, KeyChord } from "$lib/commands/types";
import { chordsEqual, eventMatchesChord, parseChord, serializeChord } from "$lib/commands/chords";
import { getCommand, listCommands } from "$lib/commands/registry";
import { PersistedState } from "runed";

const STORAGE_KEY = storageKey("keybindings");

type OverrideMap = Partial<Record<CommandId, KeyChord | null>>;

// Stored shape: { [commandId]: <parsed serializeChord JSON> | null }. Keep identical.
const overridesSerializer = {
  serialize: (map: OverrideMap): string => {
    const serializable: Record<string, unknown> = {};
    for (const [id, chord] of Object.entries(map)) {
      if (chord === null) serializable[id] = null;
      else if (chord) serializable[id] = JSON.parse(serializeChord(chord));
    }
    return JSON.stringify(serializable);
  },
  deserialize: (raw: string): OverrideMap => {
    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>;
      const out: OverrideMap = {};
      for (const [id, value] of Object.entries(parsed)) {
        if (value === null) {
          out[id as CommandId] = null;
          continue;
        }
        const chord = parseChord(value);
        if (chord) out[id as CommandId] = chord;
      }
      return out;
    } catch {
      return {};
    }
  },
};

class KeybindingsStore {
  #persisted = new PersistedState<OverrideMap>(
    STORAGE_KEY,
    {},
    {
      serializer: overridesSerializer,
    },
  );

  get overrides(): OverrideMap {
    return this.#persisted.current;
  }

  set overrides(overrides: OverrideMap) {
    this.#persisted.current = overrides;
  }

  bindingFor(id: CommandId): KeyChord | null {
    if (Object.prototype.hasOwnProperty.call(this.overrides, id)) {
      return this.overrides[id] ?? null;
    }
    return getCommand(id)?.defaultBinding ?? null;
  }

  isCustom(id: CommandId): boolean {
    return Object.prototype.hasOwnProperty.call(this.overrides, id);
  }

  setBinding(id: CommandId, chord: KeyChord | null) {
    this.overrides = { ...this.overrides, [id]: chord };
  }

  reset(id: CommandId) {
    const next = { ...this.overrides };
    delete next[id];
    this.overrides = next;
  }

  resetAll() {
    // Writes the empty map back instead of removing the key; equivalent reset semantics.
    this.overrides = {};
  }

  /** Find another command that already uses this chord (shell scope). */
  conflictFor(chord: KeyChord, except?: CommandId): CommandId | null {
    for (const cmd of listCommands()) {
      if (cmd.scope === "reader") continue;
      if (except && cmd.id === except) continue;
      const binding = this.bindingFor(cmd.id);
      if (binding && chordsEqual(binding, chord)) return cmd.id;
    }
    return null;
  }

  matchEvent(event: KeyboardEvent): CommandId | null {
    for (const cmd of listCommands()) {
      if (cmd.scope === "reader") continue;
      const binding = this.bindingFor(cmd.id);
      if (!binding) continue;
      if (eventMatchesChord(event, binding)) return cmd.id;
    }
    return null;
  }
}

export const keybindings = new KeybindingsStore();
