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
directory = "../tls"
ca_dir_url = "<CA_DIR_URL>"
ca_insecure = true

[telemetry]
enabled = true

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
	assert.Equal(t, "../tls", cfg.TLS.LetsEncrypt.Directory)
	assert.Equal(t, "<CA_DIR_URL>", cfg.TLS.LetsEncrypt.CADirURL)
	assert.Equal(t, true, cfg.TLS.LetsEncrypt.CAInsecure)

	// Telemetry
	assert.Equal(t, true, cfg.Telemetry.Enabled)

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
				"SONAR_TLS_LETSENCRYPT_DIRECTORY=../tls",
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
	assert.Equal(t, "../tls", cfg.TLS.LetsEncrypt.Directory)
	assert.Equal(t, "<CA_DIR_URL>", cfg.TLS.LetsEncrypt.CADirURL)
	assert.Equal(t, true, cfg.TLS.LetsEncrypt.CAInsecure)

	// Telemetry
	assert.Equal(t, true, cfg.Telemetry.Enabled)

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
