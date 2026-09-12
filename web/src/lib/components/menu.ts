import type { Component } from "svelte";

export interface MenuItem {
  id: string;
  label: string;
  hint?: string;
  active?: boolean;
  danger?: boolean;
  disabled?: boolean;
  separator?: boolean;
  icon?: Component<{ size?: number; class?: string }>;
  onclick?: () => void;
}
