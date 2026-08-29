package models

const (
	PermRead          int64 = 1 << 0
	PermEditMetadata  int64 = 1 << 1
	PermDeleteBooks   int64 = 1 << 2
	PermManageLibrary int64 = 1 << 3
	PermManageUsers   int64 = 1 << 4
)

const (
	DefaultUserPermissions = PermRead | PermEditMetadata | PermDeleteBooks
	AllPermissions         = PermRead | PermEditMetadata | PermDeleteBooks | PermManageLibrary | PermManageUsers
)

var permissionOrder = []struct {
	bit  int64
	name string
}{
	{PermRead, "read"},
	{PermEditMetadata, "edit_metadata"},
	{PermDeleteBooks, "delete_books"},
	{PermManageLibrary, "manage_library"},
	{PermManageUsers, "manage_users"},
}

// PermissionList returns string permission ids set in mask.
func PermissionList(mask int64) []string {
	out := make([]string, 0, len(permissionOrder))
	for _, p := range permissionOrder {
		if mask&p.bit != 0 {
			out = append(out, p.name)
		}
	}
	return out
}

// ParsePermissions builds a mask from string permission ids.
func ParsePermissions(names []string) int64 {
	var mask int64
	for _, name := range names {
		for _, p := range permissionOrder {
			if p.name == name {
				mask |= p.bit
				break
			}
		}
	}
	return mask
}

// HasPermission reports whether mask includes perm.
func HasPermission(mask, perm int64) bool {
	return mask&perm != 0
}

// EffectivePermissions returns the permission mask for a user.
func EffectivePermissions(u User) int64 {
	if u.IsAdmin {
		return AllPermissions
	}
	if u.Permissions != 0 {
		return u.Permissions
	}
	return DefaultUserPermissions
}
