---
sidebar_position: 1
---

# Server

Sonar is **self-hosted only**, so to get started you must install it to your own server.

## Prerequisites

To install the Sonar server, you must have:

- A Linux server with a public IP address (`<PUBLIC_IP>`) with Docker and Docker Compose installed.
- A registered domain name (`<DOMAIN>`).

## DNS configuration

In order for Sonar to work, it must be configured to act as a nameserver for `<DOMAIN>`.
To do this, go to your domain registrar's control panel and add a new nameserver with the domain
name `ns.<DOMAIN>` and the IP address `<PUBLIC_IP>`, then add an `NS` record for your `<DOMAIN>` pointing to
the created nameserver, i.e. `ns.<DOMAIN>`.

:::info

Let's say you have the domain `example.com` and a server with the public IP address `123.123.123.123`.
In the registrar's control panel you must add a new nameserver `ns.example.com` with IP address `123.123.123.123`.
Then, you need to add an `NS` record for `example.com`:

| example.com | record type | value          |
| ----------- | ----------- | -------------- |
| @           | NS          | ns.example.com |

:::

To ensure that everything is configured correctly, you can use the following commands:

```shell-session
$ host -t ns <DOMAIN>
<DOMAIN> name server ns.<DOMAIN>.

$ host -t a ns.<DOMAIN>
ns.<DOMAIN> has address <PUBLIC_IP>
```

## Docker compose

The recommended way to install the Sonar backend on your server is to use a Docker Compose file.
Create a `docker-compose.yml` file on your server with the following content:

```yml title="docker-compose.yml"
services:
  sonar:
    restart: always
    image: ghcr.io/nt0xa/sonar:1
    ports:
      - 21:21 # FTP
      - 25:25 # SMTP
      - 53:53/udp # DNS
      - 80:80 # HTTP
      - 443:443 # HTTPS
      - 31337:31337 # REST API
      - 31338:31338 # Webhooks (currently only used by Lark messenger in "webhook" mode)
    volumes:
      - ./tls:/opt/app/tls # TLS certificates persistance
      - ./config.toml:/opt/app/config.toml # Config file, see "Configuration"
      - ./geoip:/opt/app/geoip:ro # Directory with GeoLite2 databases, see "Configuration->geoip"

  db:
    image: postgres:16
    restart: always
    environment:
      POSTGRES_USER: sonar
      POSTGRES_PASSWORD: <POSTGRES_PASSWORD> # Put any random password here
      POSTGRES_DB: sonar
    volumes:
      - ./postgres:/var/lib/postgresql/data # Database persistance
```

## Configuration file

To configure the Sonar backend create a `config.toml` file in the same directory as the `docker-compose.yml` file.

```toml title="config.toml"
# Your server public IP address.
ip = "<PUBLIC_IP>"

# Your configured domain name.
domain = "<DOMAIN>"

[db]
# Database connection string. Use values from `docker-compose.yml`.
dsn = "postgres://sonar:<POSTGRES_PASSWORD>@db:5432/sonar?sslmode=disable"

[tls]
# Two types are currently supported: "letsencrypt" and "custom".
# If you don't have any certificates, just use "letsencrypt" and certificates will be issued
# and renewed automatically.
# Use "custom" only if you already have a long-lived wildcard TLS certificate for your domain and
# you want to use it.
type = "letsencrypt"

# Let's Encrypt TLS configuration.
[tls.letsencrypt]
# Email address, will be used for Let's Encrypt registration.
email = "<EMAIL>"

# Custom TLS configuration.
# [tls.custom]
# cert = "/path/to/cert.pem"
# key = "/path/to/key.pem"

[telemetry]
# OTEL telemetry is disabled by default.
# If enabled, the OTEL_EXPORTER_OTLP_ENDPOINT must be set.
enabled = false

[geoip]
# GeoIP lookup for IP addresses. Disabled by default.
# Requires GeoLite2 databases: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data/
enabled = true
# Path to file with City database.
city = "geoip/GeoLite2-City.mmdb"
# Path to file with ASN database.
asn = "geoip/GeoLite2-ASN.mmdb"

[modules]
# List of enabled modules. Currently three modules are supported: "api", "telegram" and "lark".
#
# "api" — provides REST API on port 31337, recommended to enable, required if you want to use CLI.
#
# Other modules provide various messengers integrations:
#
# "telegram" — Telegram messenger.
# "lark" — Lark messenger.
#
# Messenger modules can work together, but it is recommended to use only one.
# Otherwise all notifications will be sent to all the messengers.
enabled = ["api", "telegram", "lark"]

# API configuration.
[modules.api]
# Admin user token. Generate a random one using the command: `openssl rand -hex 16`
admin = "<TOKEN>"

# Telegram configuration.
[modules.telegram]
# Admin user Telegram ID. Use @getmyid_bot to get yours.
admin = "<USER_ID>"
# Bot token. Use @BotFather bot to get one.
token = "<BOT_TOKEN>"

# Lark configuration.
[modules.lark]
admin = "<ADMIN_ID>"
# App ID. You can find it on the "Credentials & Basic Info" page of your app.
app_id = "<APP_ID>"
# App Secret. You can find it on the "Credentials & Basic Info" page of your app.
app_secret = "<APP_SECRET>"
# Mode. There are two moded supported "webhook" (default) and "websocket".
# See https://open.larkoffice.com/document/uAjLw4CM/ukTMukTMukTM/server-side-sdk/golang-sdk-guide/handle-events
# for more information.
mode = "webhook"
# Verification token. Required only for "webhook" mode. You can find it on the "Events & callbacks" page of your app
# under the "Encryption strategy" tab.
verification_token = "<VERIFICATION_TOKEN>"
```

## Startup

To start the Sonar server, once you have created `docker-compose.yml` and `config.toml`, simply run:

```sh
docker compose up -d
```
