package config

import (
	"log"
	"os"
	"regexp"
	"strings"
)

// Placeholder secrets shipped in .env.example; refusing them at startup keeps
// a copy-pasted example config from ever running with a publicly known key.
const exampleJWTSecret = "change-this-to-a-random-secret"

var (
	JWTSecret     []byte
	MailboxDomain string
	CookieSecure  bool
	CORSOrigins   []string
	WebhookSecret string // Resend (Svix) signing secret; empty disables the webhook
	SetupToken    string // required to create the first admin; empty disables setup
	StaticDir     string
	TrustedProxy  []string
)

var domainRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

// Load reads configuration from the environment and exits if anything
// security-relevant is missing or unsafe.
func Load() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" || secret == exampleJWTSecret || len(secret) < 32 {
		log.Fatal("JWT_SECRET must be set to a random value of at least 32 characters (e.g. `openssl rand -hex 32`)")
	}
	JWTSecret = []byte(secret)

	MailboxDomain = strings.ToLower(strings.TrimSpace(os.Getenv("MAILBOX_DOMAIN")))
	if !domainRe.MatchString(MailboxDomain) {
		log.Fatal("MAILBOX_DOMAIN must be set to a valid domain (e.g. example.com)")
	}

	if os.Getenv("RESEND_API_KEY") == "" {
		log.Fatal("RESEND_API_KEY not set")
	}

	// Secure cookies by default; set COOKIE_SECURE=false only for plain-HTTP local dev.
	CookieSecure = os.Getenv("COOKIE_SECURE") != "false"
	CORSOrigins = splitList(os.Getenv("CORS_ORIGINS"))
	TrustedProxy = splitList(os.Getenv("TRUSTED_PROXIES"))
	WebhookSecret = os.Getenv("RESEND_WEBHOOK_SECRET")
	SetupToken = os.Getenv("SETUP_TOKEN")

	StaticDir = os.Getenv("STATIC_DIR")
	if StaticDir == "" {
		StaticDir = "../web/dist"
	}

	if !CookieSecure {
		log.Println("WARNING: COOKIE_SECURE=false, auth cookie will be sent over plain HTTP")
	}
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
