package handlers

import (
	"crypto/subtle"
	"encoding/hex"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/text/unicode/norm"
)

// Users imported from the v1 app carry Better Auth scrypt hashes
// ("<32 hex salt>:<128 hex key>"). They are verified here and upgraded to
// bcrypt on the next successful login.
var legacyScryptRe = regexp.MustCompile(`^[0-9a-f]{32}:[0-9a-f]{128}$`)

// checkPassword reports whether password matches hash, and whether the hash
// is a legacy format that should be replaced with bcrypt.
func checkPassword(hash, password string) (ok, legacy bool) {
	if legacyScryptRe.MatchString(hash) {
		return checkLegacyScrypt(hash, password), true
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil, false
}

// checkLegacyScrypt mirrors better-auth/dist/crypto/password.mjs: the hex salt
// string itself (not its decoded bytes) is the scrypt salt.
func checkLegacyScrypt(hash, password string) bool {
	salt, keyHex, _ := strings.Cut(hash, ":")
	want, err := hex.DecodeString(keyHex)
	if err != nil {
		return false
	}
	got, err := scrypt.Key([]byte(norm.NFKC.String(password)), []byte(salt), 16384, 16, 1, 64)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}
