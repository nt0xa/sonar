package sonar_test

import (
	"testing"
	"testing/fstest"

	"github.com/nt0xa/sonar/internal/cmd/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_TOML(t *testing.T) {

	testFS := fstest.MapFS{
		"config.toml": {Data: []byte(`
ip = "127.0.0.1"
domain = "example.com"

[db]
dsn = "<DB_DSN>"

[dns]
zone = "<ZONE_FILE>"

[tls]
type = "letsencrypt"

[tls.letsencrypt]
email = "<EMAIL>"
directory = "."
ca_dir_url = "<CA_DIR_URL>"
ca_insecure = true

[telemetry]
enabled = true

[audit]
enabled = true

[ratelimit]
enabled = true
allow = ["10.0.0.0/8", "2001:db8::/32"]

[ratelimit.http]
rate = 10
burst = 20

[ratelimit.smtp]
rate = 0.5
burst = 5

[modules]
enabled = ["api", "telegram", "lark"]

[modules.api]
admin = "<TOKEN>"

[modules.telegram]
admin = 1337
token = "<BOT_TOKEN>"

[modules.lark]
admin = "<ADMIN_ID>"
app_id = "<APP_ID>"
app_secret = "<APP_SECRET>"
encrypt_key = "<KEY>"
mode = "webhook"
verification_token = "<VERIFICATION_TOKEN>"
`)},
	}

	cfg, err := server.LoadConfig(
		testFS,
		func() []string { return nil },
	)
	require.NoError(t, err)

	// Basic
	assert.Equal(t, "127.0.0.1", cfg.IP)
	assert.Equal(t, "example.com", cfg.Domain)

	// DB
	assert.Equal(t, "<DB_DSN>", cfg.DB.DSN)

	// DNS
	assert.Equal(t, "<ZONE_FILE>", cfg.DNS.Zone)

	// TLS
	assert.Equal(t, "letsencrypt", cfg.TLS.Type)
	assert.Equal(t, "<EMAIL>", cfg.TLS.LetsEncrypt.Email)
	assert.Equal(t, ".", cfg.TLS.LetsEncrypt.Directory)
	assert.Equal(t, "<CA_DIR_URL>", cfg.TLS.LetsEncrypt.CADirURL)
	assert.Equal(t, true, cfg.TLS.LetsEncrypt.CAInsecure)

	// Telemetry
	assert.Equal(t, true, cfg.Telemetry.Enabled)
	assert.Equal(t, true, cfg.Audit.Enabled)

	// Rate limit
	assert.Equal(t, true, cfg.RateLimit.Enabled)
	assert.Equal(t, []string{"10.0.0.0/8", "2001:db8::/32"}, cfg.RateLimit.Allow)
	assert.Equal(t, 10.0, cfg.RateLimit.HTTP.Rate)
	assert.Equal(t, 20, cfg.RateLimit.HTTP.Burst)
	assert.Equal(t, 0.5, cfg.RateLimit.SMTP.Rate)
	assert.Equal(t, 5, cfg.RateLimit.SMTP.Burst)
	assert.Zero(t, cfg.RateLimit.FTP.Rate)

	// Test Modules config
	assert.ElementsMatch(t, []string{
		"api",
		"telegram",
		"lark",
	}, cfg.Modules.Enabled)

	// Test API
	assert.Equal(t, "<TOKEN>", cfg.Modules.API.Admin)

	// Test Telegram module config
	assert.EqualValues(t, 1337, cfg.Modules.Telegram.Admin)
	assert.Equal(t, "<BOT_TOKEN>", cfg.Modules.Telegram.Token)

	// Test Lark module config
	assert.Equal(t, "<ADMIN_ID>", cfg.Modules.Lark.Admin)
	assert.Equal(t, "<APP_ID>", cfg.Modules.Lark.AppID)
	assert.Equal(t, "<APP_SECRET>", cfg.Modules.Lark.AppSecret)
	assert.Equal(t, "<KEY>", cfg.Modules.Lark.EncryptKey)
	assert.Equal(t, "webhook", cfg.Modules.Lark.Mode)
	assert.Equal(t, "<VERIFICATION_TOKEN>", cfg.Modules.Lark.VerificationToken)
}

func TestConfig_Invalid(t *testing.T) {
	_, err := server.LoadConfig(
		fstest.MapFS{},
		func() []string { return nil },
	)
	require.Error(t, err)
	// Missing required domain/ip surface as keyed problems in the message.
	assert.Contains(t, err.Error(), "domain")
	assert.Contains(t, err.Error(), "ip")
}

func TestConfig_Env(t *testing.T) {
	cfg, err := server.LoadConfig(
		fstest.MapFS{},
		func() []string {
			return []string{
				"SONAR_IP=127.0.0.1",
				"SONAR_DOMAIN=example.com",
				"SONAR_DB_DSN=<DB_DSN>",
				"SONAR_DNS_ZONE=<ZONE_FILE>",
				"SONAR_TLS_TYPE=letsencrypt",
				"SONAR_TLS_LETSENCRYPT_EMAIL=<EMAIL>",
				"SONAR_TLS_LETSENCRYPT_DIRECTORY=.",
				"SONAR_TLS_LETSENCRYPT_CA_DIR_URL=<CA_DIR_URL>",
				"SONAR_TLS_LETSENCRYPT_CA_INSECURE=true",
				"SONAR_MODULES_ENABLED=api,telegram,lark",
				"SONAR_MODULES_API_ADMIN=<TOKEN>",
				"SONAR_MODULES_TELEGRAM_ADMIN=1337",
				"SONAR_MODULES_TELEGRAM_TOKEN=<BOT_TOKEN>",
				"SONAR_MODULES_LARK_ADMIN=<ADMIN_ID>",
				"SONAR_MODULES_LARK_MODE=webhook",
				"SONAR_MODULES_LARK_APP_ID=<APP_ID>",
				"SONAR_MODULES_LARK_APP_SECRET=<APP_SECRET>",
				"SONAR_MODULES_LARK_ENCRYPT_KEY=<KEY>",
				"SONAR_MODULES_LARK_VERIFICATION_TOKEN=<VERIFICATION_TOKEN>",
				"SONAR_TELEMETRY_ENABLED=true",
				"SONAR_AUDIT_ENABLED=true",
				"SONAR_RATELIMIT_ENABLED=true",
				"SONAR_RATELIMIT_ALLOW=10.0.0.0/8,2001:db8::/32",
				"SONAR_RATELIMIT_HTTP_RATE=10",
				"SONAR_RATELIMIT_HTTP_BURST=20",
				"SONAR_RATELIMIT_SMTP_RATE=0.5",
				"SONAR_RATELIMIT_SMTP_BURST=5",
			}
		},
	)
	require.NoError(t, err)

	// Basic
	assert.Equal(t, "127.0.0.1", cfg.IP)
	assert.Equal(t, "example.com", cfg.Domain)

	// DB
	assert.Equal(t, "<DB_DSN>", cfg.DB.DSN)

	// DNS
	assert.Equal(t, "<ZONE_FILE>", cfg.DNS.Zone)

	// TLS
	assert.Equal(t, "letsencrypt", cfg.TLS.Type)
	assert.Equal(t, "<EMAIL>", cfg.TLS.LetsEncrypt.Email)
	assert.Equal(t, ".", cfg.TLS.LetsEncrypt.Directory)
	assert.Equal(t, "<CA_DIR_URL>", cfg.TLS.LetsEncrypt.CADirURL)
	assert.Equal(t, true, cfg.TLS.LetsEncrypt.CAInsecure)

	// Telemetry
	assert.Equal(t, true, cfg.Telemetry.Enabled)
	assert.Equal(t, true, cfg.Audit.Enabled)

	// Rate limit
	assert.Equal(t, true, cfg.RateLimit.Enabled)
	assert.Equal(t, []string{"10.0.0.0/8", "2001:db8::/32"}, cfg.RateLimit.Allow)
	assert.Equal(t, 10.0, cfg.RateLimit.HTTP.Rate)
	assert.Equal(t, 20, cfg.RateLimit.HTTP.Burst)
	assert.Equal(t, 0.5, cfg.RateLimit.SMTP.Rate)
	assert.Equal(t, 5, cfg.RateLimit.SMTP.Burst)
	assert.Zero(t, cfg.RateLimit.FTP.Rate)

	// Test Modules config
	assert.ElementsMatch(t, []string{
		"api",
		"telegram",
		"lark",
	}, cfg.Modules.Enabled)

	// Test API
	assert.Equal(t, "<TOKEN>", cfg.Modules.API.Admin)

	// Test Telegram module config
	assert.EqualValues(t, 1337, cfg.Modules.Telegram.Admin)
	assert.Equal(t, "<BOT_TOKEN>", cfg.Modules.Telegram.Token)

	// Test Lark module config
	assert.Equal(t, "<ADMIN_ID>", cfg.Modules.Lark.Admin)
	assert.Equal(t, "<APP_ID>", cfg.Modules.Lark.AppID)
	assert.Equal(t, "<APP_SECRET>", cfg.Modules.Lark.AppSecret)
	assert.Equal(t, "<KEY>", cfg.Modules.Lark.EncryptKey)
	assert.Equal(t, "webhook", cfg.Modules.Lark.Mode)
	assert.Equal(t, "<VERIFICATION_TOKEN>", cfg.Modules.Lark.VerificationToken)
}

func TestConfig_InvalidRateLimit(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		err  string
	}{
		{"bad allow", []string{"SONAR_RATELIMIT_ALLOW=10.0.0.1"}, "ratelimit.allow: element #0: must be a valid CIDR prefix"},
		{"negative rate", []string{"SONAR_RATELIMIT_HTTP_RATE=-1", "SONAR_RATELIMIT_HTTP_BURST=1"}, "ratelimit.http.rate: must be >= 0"},
		{"missing burst", []string{"SONAR_RATELIMIT_SMTP_RATE=1"}, "ratelimit.smtp.burst: must be >= 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadRateLimitConfig(append(tt.env, "SONAR_RATELIMIT_ENABLED=true"))
			require.Error(t, err)
			assert.Equal(t, "validation failed: "+tt.err, err.Error())
		})
	}
}

func TestConfig_DisabledRateLimitIsNotValidated(t *testing.T) {
	cfg, err := loadRateLimitConfig([]string{"SONAR_RATELIMIT_ALLOW=10.0.0.1"})
	require.NoError(t, err)
	assert.False(t, cfg.RateLimit.Enabled)
}

func loadRateLimitConfig(env []string) (*server.Config, error) {
	return server.LoadConfig(
		fstest.MapFS{},
		func() []string {
			return append([]string{
				"SONAR_IP=127.0.0.1",
				"SONAR_DOMAIN=example.com",
				"SONAR_DB_DSN=<DB_DSN>",
				"SONAR_TLS_TYPE=letsencrypt",
				"SONAR_TLS_LETSENCRYPT_EMAIL=<EMAIL>",
				"SONAR_TLS_LETSENCRYPT_DIRECTORY=.",
				"SONAR_MODULES_API_ADMIN=<TOKEN>",
			}, env...)
		},
	)
}
