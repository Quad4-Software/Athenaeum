package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2idPrefix = "$argon2id$"
	argon2Version  = argon2.Version
	argon2KeyLen   = 32
	argon2SaltLen  = 16

	// RFC 9106 second option adapted for self-hosted boxes (single thread).
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // KiB (64 MiB)
	argon2Threads = 1

	// Tiny params keep -race and unit tests fast.
	argon2TestTime    = 1
	argon2TestMemory  = 8 // KiB (argon2 minimum with p=1)
	argon2TestThreads = 1
)

type argon2Params struct {
	time    uint32
	memory  uint32
	threads uint8
}

func activeArgon2Params() argon2Params {
	if testing.Testing() {
		return argon2Params{
			time:    argon2TestTime,
			memory:  argon2TestMemory,
			threads: argon2TestThreads,
		}
	}
	return argon2Params{
		time:    argon2Time,
		memory:  argon2Memory,
		threads: argon2Threads,
	}
}

// HashPassword returns an argon2id PHC-encoded hash of password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	p := activeArgon2Params()
	sum := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, argon2KeyLen)
	return encodeArgon2id(p, salt, sum), nil
}

// CheckPassword reports whether password matches hash.
// Accepts argon2id PHC strings and legacy bcrypt hashes.
func CheckPassword(hash, password string) bool {
	switch {
	case strings.HasPrefix(hash, argon2idPrefix):
		return checkArgon2id(hash, password)
	case strings.HasPrefix(hash, "$2a$"), strings.HasPrefix(hash, "$2b$"), strings.HasPrefix(hash, "$2y$"):
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	default:
		return false
	}
}

// NeedsRehash reports whether hash should be replaced with current argon2id params.
// True for legacy bcrypt and for argon2id encoded with different parameters.
func NeedsRehash(hash string) bool {
	if !strings.HasPrefix(hash, argon2idPrefix) {
		return true
	}
	parsed, err := parseArgon2id(hash)
	if err != nil {
		return true
	}
	want := activeArgon2Params()
	return parsed.time != want.time || parsed.memory != want.memory || parsed.threads != want.threads ||
		parsed.version != argon2Version || len(parsed.sum) != argon2KeyLen
}

func encodeArgon2id(p argon2Params, salt, sum []byte) string {
	b64 := base64.RawStdEncoding
	return fmt.Sprintf(
		"%sv=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idPrefix,
		argon2Version,
		p.memory,
		p.time,
		p.threads,
		b64.EncodeToString(salt),
		b64.EncodeToString(sum),
	)
}

type parsedArgon2id struct {
	version uint32
	time    uint32
	memory  uint32
	threads uint8
	salt    []byte
	sum     []byte
}

func parseArgon2id(encoded string) (parsedArgon2id, error) {
	// $argon2id$v=19$m=65536,t=3,p=1$salt$hash
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return parsedArgon2id{}, fmt.Errorf("invalid argon2id encoding")
	}
	var out parsedArgon2id
	if _, err := fmt.Sscanf(parts[2], "v=%d", &out.version); err != nil {
		return parsedArgon2id{}, fmt.Errorf("invalid argon2id version")
	}
	for part := range strings.SplitSeq(parts[3], ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return parsedArgon2id{}, fmt.Errorf("invalid argon2id params")
		}
		n, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return parsedArgon2id{}, fmt.Errorf("invalid argon2id params")
		}
		switch kv[0] {
		case "m":
			out.memory = uint32(n)
		case "t":
			out.time = uint32(n)
		case "p":
			if n == 0 || n > 255 {
				return parsedArgon2id{}, fmt.Errorf("invalid argon2id parallelism")
			}
			out.threads = uint8(n) // #nosec G115 -- bounded to 1..255 above
		default:
			return parsedArgon2id{}, fmt.Errorf("unknown argon2id param %q", kv[0])
		}
	}
	if out.memory == 0 || out.time == 0 || out.threads == 0 {
		return parsedArgon2id{}, fmt.Errorf("incomplete argon2id params")
	}
	b64 := base64.RawStdEncoding
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return parsedArgon2id{}, err
	}
	sum, err := b64.DecodeString(parts[5])
	if err != nil {
		return parsedArgon2id{}, err
	}
	out.salt = salt
	out.sum = sum
	return out, nil
}

func checkArgon2id(encoded, password string) bool {
	parsed, err := parseArgon2id(encoded)
	if err != nil {
		return false
	}
	if parsed.version != argon2Version {
		return false
	}
	keyLen := len(parsed.sum)
	if keyLen == 0 || keyLen > 1024 {
		return false
	}
	sum := argon2.IDKey(
		[]byte(password),
		parsed.salt,
		parsed.time,
		parsed.memory,
		parsed.threads,
		uint32(keyLen), // #nosec G115 -- bounded to 1..1024 above
	)
	return subtle.ConstantTimeCompare(sum, parsed.sum) == 1
}
