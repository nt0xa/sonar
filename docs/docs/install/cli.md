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
