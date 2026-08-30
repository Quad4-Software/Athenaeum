package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/config"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
	"athenaeum/internal/term"
)

// RunUsers dispatches `athenaeum users` subcommands.
func RunUsers(args []string) error {
	applyUsersColor(args)
	configurePasswordPolicyFromEnv()
	if len(args) == 0 {
		printUsersHelp(os.Stdout)
		return nil
	}
	switch args[0] {
	case "list", "ls":
		return usersList(args[1:])
	case "add", "create":
		return usersAdd(args[1:])
	case "reset-password", "passwd", "password":
		return usersResetPassword(args[1:])
	case "rename":
		return usersRename(args[1:])
	case "set-admin", "admin":
		return usersSetAdmin(args[1:])
	case "set-permissions", "permissions":
		return usersSetPermissions(args[1:])
	case "delete", "rm":
		return usersDelete(args[1:])
	case "show", "get":
		return usersShow(args[1:])
	case "help", "-h", "--help":
		printUsersHelp(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown users command %q (run '%s users help')", args[0], binaryName())
	}
}

func applyUsersColor(args []string) {
	mode := term.ModeAuto
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--no-color":
			mode = term.ModeNever
		case a == "--color" && i+1 < len(args):
			if m, err := term.ParseMode(args[i+1]); err == nil {
				mode = m
			}
			i++
		case strings.HasPrefix(a, "--color="):
			if m, err := term.ParseMode(strings.TrimPrefix(a, "--color=")); err == nil {
				mode = m
			}
		}
	}
	if config.EnvBool("ATHENAEUM_NO_COLOR", false) {
		mode = term.ModeNever
	}
	term.Apply(mode)
}

func printUsersHelp(w io.Writer) {
	bin := binaryName()
	term.Fprintln(w, term.Bold(w, "Manage local user accounts")+" in the "+brand.Name+" database.")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Usage:"))
	term.Fprintf(w, "  %s users <command> [flags] [args]\n", bin)
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Commands:"))
	printCmd(w, "list, ls", "List users")
	printCmd(w, "add, create <username>", "Create a local user")
	printCmd(w, "reset-password <user>", "Reset a user's password (passwd, password)")
	printCmd(w, "rename <user> <name>", "Rename a user")
	printCmd(w, "set-admin <user>", "Grant or revoke admin")
	printCmd(w, "set-permissions <user>", "Set non-admin permissions")
	printCmd(w, "show, get <user>", "Show one user")
	printCmd(w, "delete, rm <user>", "Delete a user")
	printCmd(w, "help", "Show this help")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Global flags:"))
	printCmd(w, "--data <dir>", "Data directory (default ./data or ATHENAEUM_DATA)")
	printCmd(w, "--color <mode>", "auto, always, or never")
	printCmd(w, "--no-color", "Disable ANSI color")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Password sources (in order):"))
	term.Fprintln(w, "  --password, ATHENAEUM_PASSWORD, or stdin")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Examples:"))
	term.Fprintf(w, "  %s\n", term.Dim(w, bin+" users list --data ./data"))
	term.Fprintf(w, "  %s\n", term.Dim(w, bin+" users add alice --password 'secretpass'"))
	term.Fprintf(w, "  %s\n", term.Dim(w, bin+" users set-admin bob --no-admin"))
}

func printCmd(w io.Writer, name, desc string) {
	term.Fprintf(w, "  %-28s %s\n", term.Flag(w, name), desc)
}

func usersList(args []string) error {
	fs := flag.NewFlagSet("users list", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	users, err := store.ListUsers(context.Background())
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, term.Bold(os.Stdout, "ID\tUSERNAME\tADMIN\tGUEST\tLOCAL\tPERMISSIONS\tCREATED"))
	for _, u := range users {
		perms := strings.Join(u.PermissionNames(), ",")
		if perms == "" {
			perms = "-"
		}
		guest := "no"
		if u.IsGuest {
			guest = "yes"
		}
		local := "no"
		if u.LocalAuth {
			local = "yes"
		}
		admin := "no"
		if u.IsAdmin {
			admin = "yes"
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			u.ID, u.Username, admin, guest, local, perms, u.CreatedAt.Format(time.RFC3339))
	}
	return tw.Flush()
}

func usersAdd(args []string) error {
	fs := flag.NewFlagSet("users add", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	password := fs.String("password", "", "password (or ATHENAEUM_PASSWORD or stdin)")
	admin := fs.Bool("admin", false, "grant admin")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users add <username> [--password PASS] [--admin]", binaryName())
	}
	username := strings.TrimSpace(fs.Arg(0))
	if len(username) < 2 {
		return errors.New("username must be at least 2 characters")
	}

	pass, err := readPassword(*password)
	if err != nil {
		return err
	}
	if err := auth.ValidatePassword(pass); err != nil {
		return err
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	taken, err := store.UsernameTaken(ctx, username, 0)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("username %q is already taken", username)
	}

	hash, err := auth.HashPassword(pass)
	if err != nil {
		return err
	}
	id, err := store.CreateUser(ctx, username, hash, *admin)
	if err != nil {
		return err
	}
	Successf("created user id=%d username=%s admin=%t", id, username, *admin)
	return nil
}

func usersResetPassword(args []string) error {
	fs := flag.NewFlagSet("users reset-password", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	password := fs.String("password", "", "new password (or ATHENAEUM_PASSWORD or stdin)")
	revoke := fs.Bool("revoke-sessions", true, "revoke active sessions after reset")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users reset-password <user> [--password PASS]", binaryName())
	}

	pass, err := readPassword(*password)
	if err != nil {
		return err
	}
	if err := auth.ValidatePassword(pass); err != nil {
		return err
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	u, err := resolveUser(ctx, store, fs.Arg(0))
	if err != nil {
		return err
	}

	hash, err := auth.HashPassword(pass)
	if err != nil {
		return err
	}
	if err := store.UpdateUserPassword(ctx, u.ID, hash); err != nil {
		return err
	}
	if *revoke {
		if _, err := store.RevokeUserSessions(ctx, u.ID); err != nil {
			return err
		}
	}
	Successf("password updated for %s (id=%d)", u.Username, u.ID)
	return nil
}

func usersRename(args []string) error {
	fs := flag.NewFlagSet("users rename", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: %s users rename <user> <new-username>", binaryName())
	}
	newName := strings.TrimSpace(fs.Arg(1))
	if len(newName) < 2 {
		return errors.New("username must be at least 2 characters")
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	u, err := resolveUser(ctx, store, fs.Arg(0))
	if err != nil {
		return err
	}
	taken, err := store.UsernameTaken(ctx, newName, u.ID)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("username %q is already taken", newName)
	}
	if err := store.UpdateUsername(ctx, u.ID, newName); err != nil {
		return err
	}
	Successf("renamed user id=%d to %q", u.ID, newName)
	return nil
}

func usersSetAdmin(args []string) error {
	fs := flag.NewFlagSet("users set-admin", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	grant := fs.Bool("admin", false, "grant admin privileges")
	noAdmin := fs.Bool("no-admin", false, "revoke admin privileges")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users set-admin <user> [--admin|--no-admin]", binaryName())
	}
	makeAdmin := true
	if *noAdmin {
		makeAdmin = false
	} else if *grant {
		makeAdmin = true
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	u, err := resolveUser(ctx, store, fs.Arg(0))
	if err != nil {
		return err
	}
	if u.IsAdmin && !makeAdmin {
		n, err := store.AdminCount(ctx)
		if err != nil {
			return err
		}
		if n <= 1 {
			return errors.New("cannot remove admin from the last admin account")
		}
	}
	if err := store.SetUserAdmin(ctx, u.ID, makeAdmin); err != nil {
		return err
	}
	Successf("user %s (id=%d) admin=%t", u.Username, u.ID, makeAdmin)
	return nil
}

func usersSetPermissions(args []string) error {
	fs := flag.NewFlagSet("users set-permissions", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	set := fs.String("set", "", "comma-separated permissions: read,edit_metadata,delete_books,manage_library,manage_users")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users set-permissions <user> --set read,edit_metadata", binaryName())
	}
	if strings.TrimSpace(*set) == "" {
		return errors.New("--set is required")
	}

	names := splitCSV(*set)
	mask := models.ParsePermissions(names)
	if mask == 0 {
		return fmt.Errorf("no valid permissions in %q", *set)
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	u, err := resolveUser(ctx, store, fs.Arg(0))
	if err != nil {
		return err
	}
	if u.IsAdmin {
		return errors.New("admin users always have all permissions; use set-admin --no-admin first")
	}
	if err := store.SetUserPermissions(ctx, u.ID, mask); err != nil {
		return err
	}
	Successf("user %s (id=%d) permissions=%s", u.Username, u.ID, strings.Join(models.PermissionList(mask), ","))
	return nil
}

func usersDelete(args []string) error {
	fs := flag.NewFlagSet("users delete", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	force := fs.Bool("force", false, "skip confirmation prompt")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users delete <user> [--force]", binaryName())
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	u, err := resolveUser(ctx, store, fs.Arg(0))
	if err != nil {
		return err
	}
	if u.IsAdmin {
		n, err := store.AdminCount(ctx)
		if err != nil {
			return err
		}
		if n <= 1 {
			return errors.New("cannot delete the last admin account")
		}
	}
	if !*force {
		fmt.Fprintf(os.Stderr, "delete user %s (id=%d)? [y/N] ", u.Username, u.ID)
		var answer string
		if _, err := fmt.Fscanln(os.Stdin, &answer); err != nil {
			return errors.New("aborted")
		}
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			return errors.New("aborted")
		}
	}
	if err := store.DeleteUser(ctx, u.ID); err != nil {
		return err
	}
	Successf("deleted user %s (id=%d)", u.Username, u.ID)
	return nil
}

func usersShow(args []string) error {
	fs := flag.NewFlagSet("users show", flag.ContinueOnError)
	dataDir := fs.String("data", config.Env("ATHENAEUM_DATA", "./data"), "data directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: %s users show <user>", binaryName())
	}

	store, err := openStore(*dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	u, err := resolveUser(context.Background(), store, fs.Arg(0))
	if err != nil {
		return err
	}
	pub := u.Public()
	fmt.Fprintf(os.Stdout, "id:          %d\n", pub.ID)
	fmt.Fprintf(os.Stdout, "username:    %s\n", pub.Username)
	if pub.Email != "" {
		fmt.Fprintf(os.Stdout, "email:       %s\n", pub.Email)
	}
	fmt.Fprintf(os.Stdout, "admin:       %t\n", pub.IsAdmin)
	fmt.Fprintf(os.Stdout, "guest:       %t\n", pub.IsGuest)
	fmt.Fprintf(os.Stdout, "local_auth:  %t\n", pub.LocalAuth)
	fmt.Fprintf(os.Stdout, "permissions: %s\n", strings.Join(pub.Permissions, ","))
	fmt.Fprintf(os.Stdout, "created:     %s\n", pub.CreatedAt.Format(time.RFC3339))
	if pub.ExpiresAt != nil {
		fmt.Fprintf(os.Stdout, "expires:     %s\n", pub.ExpiresAt.Format(time.RFC3339))
	}
	return nil
}

func openStore(dataDir string) (*storage.Store, error) {
	driver, err := storage.ParseDriver(config.Env("ATHENAEUM_DATABASE_DRIVER", "sqlite"))
	if err != nil {
		return nil, err
	}
	if driver == storage.DriverPostgres {
		url := strings.TrimSpace(config.Env("ATHENAEUM_DATABASE_URL", ""))
		if url == "" {
			return nil, errors.New("ATHENAEUM_DATABASE_URL is required when ATHENAEUM_DATABASE_DRIVER=postgres")
		}
		return storage.OpenWith(storage.OpenOptions{Driver: driver, URL: url})
	}
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	dbPath := config.ResolveDBPath(abs)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("database not found at %s (is --data correct?)", dbPath)
	}
	return storage.Open(dbPath)
}

func resolveUser(ctx context.Context, store *storage.Store, ref string) (models.User, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return models.User{}, errors.New("user reference is required")
	}
	if id, err := strconv.ParseInt(ref, 10, 64); err == nil && id > 0 {
		u, err := store.GetUser(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return models.User{}, fmt.Errorf("user id %d not found", id)
			}
			return models.User{}, err
		}
		return u, nil
	}
	u, _, err := store.GetUserByUsername(ctx, ref)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return models.User{}, fmt.Errorf("user %q not found", ref)
		}
		return models.User{}, err
	}
	return u, nil
}

func readPassword(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := strings.TrimSpace(config.Env("ATHENAEUM_PASSWORD", "")); v != "" {
		return v, nil
	}
	if stat, _ := os.Stdin.Stat(); stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		sc := bufio.NewScanner(os.Stdin)
		if sc.Scan() {
			return strings.TrimSpace(sc.Text()), nil
		}
		return "", errors.New("no password on stdin")
	}
	fmt.Fprint(os.Stderr, "Password: ")
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return "", errors.New("no password provided")
	}
	return strings.TrimSpace(sc.Text()), sc.Err()
}

func configurePasswordPolicyFromEnv() {
	minLength, _ := strconv.Atoi(config.Env("ATHENAEUM_PASSWORD_MIN_LENGTH", "8"))
	longLength, _ := strconv.Atoi(config.Env("ATHENAEUM_PASSWORD_LONG_LENGTH", "12"))
	minKinds, _ := strconv.Atoi(config.Env("ATHENAEUM_PASSWORD_MIN_KINDS", "3"))
	auth.SetPasswordPolicy(auth.PasswordPolicy{
		MinLength:     minLength,
		LongLength:    longLength,
		MinKinds:      minKinds,
		RequireLower:  config.EnvBool("ATHENAEUM_PASSWORD_REQUIRE_LOWER", false),
		RequireUpper:  config.EnvBool("ATHENAEUM_PASSWORD_REQUIRE_UPPER", false),
		RequireDigit:  config.EnvBool("ATHENAEUM_PASSWORD_REQUIRE_DIGIT", false),
		RequireSymbol: config.EnvBool("ATHENAEUM_PASSWORD_REQUIRE_SYMBOL", false),
	})
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
