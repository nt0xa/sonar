---
sidebar_position: 2
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# CLI

:::warning

The CLI client uses the REST API, so in order to be able to use it, make sure that the "api" module
is enabled in your Sonar server's `config.toml` file.

See [Server: Configuration file](/sonar/install/server#configuration-file)

:::

## Installation


<Tabs>
<TabItem value="macOS" label="macOS" default>
```shell-session
$ brew install nt0xa/sonar/sonar
```
</TabItem>
<TabItem value="linux" label="Linux">
Download binaries for the latest release from [Github](https://github.com/nt0xa/sonar/releases).
</TabItem>
<TabItem value="windows" label="Windows">
Download binaries for the latest release from [Github](https://github.com/nt0xa/sonar/releases).
</TabItem>
</Tabs>


## API token


## Configration file

To start using the CLI, you must first create the configuration file at `~/.config/sonar/config.toml`.
To configure sever you only need two values:

- `<DOMAIN>` — your server's domain.
- `<TOKEN>` — your user's token. If you are the one who deployed the server, you can use
  the token from the [Server: Configuration file](/sonar/install/server#configuration-file). Otherwise,
  you can go to the configured messenger and use the `/profile` command to get your token.

  ![Getting token in Telegram](../assets/telegram_token_dark.png#gh-dark-mode-only)![Getting token in Telegram](../assets/telegram_token_light.png#gh-light-mode-only)


Here is an example configuration:

```toml title="~/.config/sonar/config.toml"
[servers]
[servers.myserver1]
token = "<TOKEN>"
url = "https://<DOMAIN>:31337"

# You can add another server here, if you have more than one.
# [servers.myserver2]
# token = "<TOKEN2>"
# url = "https://<DOMAIN2>:31337"

[context]
# The server that is currently active.
server = "myserver1"
```

## MCP server

`sonar mcp` starts an [MCP](https://modelcontextprotocol.io) server over stdio, so AI assistants can manage
payloads, DNS records and HTTP routes and read events. It uses the same configuration file as the CLI
and talks to the server from the current context.

Add it to your MCP client, for example Claude Code:

```shell-session
$ claude mcp add sonar -- sonar mcp
```

Or any client that uses the common JSON config format:

```json
{
  "mcpServers": {
    "sonar": {
      "command": "sonar",
      "args": ["mcp"]
    }
  }
}
```

To use a different configuration file, pass `--config`: `sonar mcp --config /path/to/config.toml`.

Available tools:

| Tool | Description |
| --- | --- |
| `payloads_list`, `payloads_create`, `payloads_update`, `payloads_delete`, `payloads_clear` | Manage payloads |
| `dns_records_list`, `dns_records_create`, `dns_records_delete`, `dns_records_clear` | Manage DNS records |
| `http_routes_list`, `http_routes_create`, `http_routes_update`, `http_routes_delete`, `http_routes_clear` | Manage HTTP routes |
| `events_list`, `events_get` | View events |

Notes:

- Admin commands (users and audit records) are not available over MCP.
- Payloads created over MCP don't send messenger notifications unless `notifyProtocols` is set.
- In results, `subdomain` and `payloadSubdomain` contain the full payload hostname (e.g. `abc123.sonar.example.com`), built from the server URL in your config.
- Bodies and raw event data are returned as text, or as a hexdump if they are binary. `events_list` omits raw data; use `events_get` for it.
