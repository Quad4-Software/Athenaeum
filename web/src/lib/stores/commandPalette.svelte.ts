/**
 * Open/close state for the global command palette.
 */

class CommandPaletteStore {
  open = $state(false);
  query = $state("");

  show(query = "") {
    this.query = query;
    this.open = true;
  }

  hide() {
    this.open = false;
    this.query = "";
  }

  toggle() {
    if (this.open) this.hide();
    else this.show();
  }
}

export const commandPalette = new CommandPaletteStore();
