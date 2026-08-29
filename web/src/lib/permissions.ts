import type { Permission } from "$lib/api/types";
import { auth } from "$lib/stores/auth.svelte";

export const PERMISSION_LABELS: Record<Permission, string> = {
  read: "Browse library",
  edit_metadata: "Edit metadata",
  delete_books: "Delete books",
  manage_library: "Manage libraries",
  manage_users: "Manage users",
};

export const EDITABLE_PERMISSIONS: Permission[] = [
  "read",
  "edit_metadata",
  "delete_books",
  "manage_library",
  "manage_users",
];

export function can(permission: Permission): boolean {
  if (!auth.authEnabled) return true;
  const user = auth.user;
  if (!user) return false;
  if (user.isAdmin) return true;
  return user.permissions?.includes(permission) ?? false;
}
