package sonar_test

import (
	"fmt"
	"testing"

	"github.com/nt0xa/sonar/internal/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := server.GetConfig(
		map[string]any{
			"ip": "<DEFAULT_IP>",
		},
		[]byte(`
ip = "<CONFIG_IP>"
domain = "<DOMAIN>"

[db]
dsn = "<DB_DSN>"

[tls]
type = "letsencrypt"

[tls.letsencrypt]
email = "<EMAIL>"

[modules]
enabled = ["api", "telegram", "lark"]

[modules.api]
admin = "<TOKEN>"

[modules.telegram]
admin = "<USER_ID>"
token = "<BOT_TOKEN>"

[modules.lark]
admin = "<ADMIN_ID>"
app_id = "<APP_ID>"
app_secret = "<APP_SECRET>"
mode = "webhook"
verification_token = "<VERIFICATION_TOKEN>"
`),
		func() []string {
			return []string{
				"SONAR_IP=<ENV_IP>",
			}
		},
	)
	require.NoError(t, err)

	// Test TLS LetsEncrypt config
	assert.Equal(t, "./tls", cfg.TLS.LetsEncrypt.Directory)
	assert.Equal(t,
		"https://acme-v02.api.letsencrypt.org/directory",
		cfg.TLS.LetsEncrypt.CADirURL)
	assert.Equal(t, "<EMAIL>", cfg.TLS.LetsEncrypt.Email)
	assert.Equal(t, "letsencrypt", cfg.TLS.Type)

	// Test basic config fields
	assert.Equal(t, "<CONFIG_IP>", cfg.IP)
	assert.Equal(t, "<DOMAIN>", cfg.Domain)

	// Test DB config
	assert.Equal(t, "<DB_DSN>", cfg.DB.DSN)

	// Test Modules config
	assert.ElementsMatch(t, []string{"api", "telegram", "lark"}, cfg.Modules.Enabled)

	// Test API module config
	assert.Equal(t, "<TOKEN>", cfg.Modules.API.Admin)

	// Test Telegram module config
	assert.Equal(t, "<USER_ID>", cfg.Modules.Telegram.Admin)
	assert.Equal(t, "<BOT_TOKEN>", cfg.Modules.Telegram.Token)

	// Test Lark module config
	assert.Equal(t, "<ADMIN_ID>", cfg.Modules.Lark.Admin)
	assert.Equal(t, "<APP_ID>", cfg.Modules.Lark.AppID)
	assert.Equal(t, "<APP_SECRET>", cfg.Modules.Lark.AppSecret)
	assert.Equal(t, "webhook", cfg.Modules.Lark.Mode)
	assert.Equal(t, "<VERIFICATION_TOKEN>", cfg.Modules.Lark.VerificationToken)

	fmt.Printf("cfg = %+v\n", cfg)
}
