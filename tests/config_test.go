package sonar_test

import (
	"testing"

	"github.com/nt0xa/sonar/internal/cmd/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_TOML(t *testing.T) {
	cfg, err := server.GetConfig(
		nil,
		[]byte(`
ip = "<IP>"
domain = "<DOMAIN>"

[db]
dsn = "<DB_DSN>"

[dns]
zone = "<ZONE_FILE>"

[tls]
type = "letsencrypt"

[tls.letsencrypt]
email = "<EMAIL>"
directory = "<DIR>"
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
mode = "webhook"
verification_token = "<VERIFICATION_TOKEN>"
`),
		func() []string { return nil },
	)
	require.NoError(t, err)

	// Basic
	assert.Equal(t, "<IP>", cfg.IP)
	assert.Equal(t, "<DOMAIN>", cfg.Domain)

	// DB
	assert.Equal(t, "<DB_DSN>", cfg.DB.DSN)

	// DNS
	assert.Equal(t, "<ZONE_FILE>", cfg.DNS.Zone)

	// TLS
	assert.Equal(t, "letsencrypt", cfg.TLS.Type)
	assert.Equal(t, "<EMAIL>", cfg.TLS.LetsEncrypt.Email)
	assert.Equal(t, "<DIR>", cfg.TLS.LetsEncrypt.Directory)
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

	// Test Telegram module config
	assert.EqualValues(t, 1337, cfg.Modules.Telegram.Admin)
	assert.Equal(t, "<BOT_TOKEN>", cfg.Modules.Telegram.Token)

	// Test Lark module config
	assert.Equal(t, "<ADMIN_ID>", cfg.Modules.Lark.Admin)
	assert.Equal(t, "<APP_ID>", cfg.Modules.Lark.AppID)
	assert.Equal(t, "<APP_SECRET>", cfg.Modules.Lark.AppSecret)
	assert.Equal(t, "webhook", cfg.Modules.Lark.Mode)
	assert.Equal(t, "<VERIFICATION_TOKEN>", cfg.Modules.Lark.VerificationToken)
}
