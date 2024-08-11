---
sidebar_position: 1
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Quick start


Sonar can be managed either by sending messages to the chat with the bot in Messenger or using the [CLI](/sonar/install/CLI).


<Tabs>
<TabItem value="messenger" label="Messenger" default>
1. Open a chat with the Sonar bot in the configured messenger.
2. Create a new payload with the command `/new <NAME>`. Select a meaningful name as it will be used in
   in interaction notifications. You will receive your unique domain name in the response.

   ![Payload creation in Telegram](../assets/create_payload_telegram_dark.png#gh-dark-mode-only)![Payload creation in Telegram](../assets/create_payload_telegram_light.png#gh-light-mode-only)
</TabItem>
<TabItem value="cli" label="CLI" default>
1. Install and configure CLI as described [here](/sonar/install/client).
2. Create a new payload with the command `sonar new <NAME>`. Select a meaningful name as it will be used in
   in interaction notifications. You will receive your unique domain name in the response.

   ![Payload creation in CLI](../assets/create_payload_cli_dark.png#gh-dark-mode-only)![Payload creation in CLI](../assets/create_payload_cli_light.png#gh-light-mode-only)
</TabItem>
</Tabs>

:::tip

- `project_test` — payload's name
- `d14a68e4.sonar.test` — payload's unique subdomain
- `dns, ftp, http, smtp` — protocols for which notifications are enabled (by default all protocols are enabled)
- `false` — shows if all interaction events are stored in database (disabled by default)

:::

3. You can now use your unique domain `d14a68e4.sonar.test` in any DNS/HTTP/SMTP/FTP interactions and
   you will receive notifications to the chat with the Sonar bot for all the interactions.
   Here is an example HTTP interaction notifications after execution of of the command `curl d14a68e4.sonar.test`:

   ![Example HTTP notification](../assets/http_notification_dark.png#gh-dark-mode-only)![Example HTTP notification](../assets/http_notification_light.png#gh-light-mode-only)


 :::tip

 - `project_test` — payload's name (the same as was used in the `/new' command when the payload was created)
 - `HTTP` — protocol of the iteraction
 - `100.100.100.100:12345` — IP address and port from which the interaction occurred
 - `04 Aug 2024 at 19:58:50 BST` — date and time of the interaction
 - The interaction details:

   ```
   GET / HTTP/1.1
   Host: d14a68e4.sonar.test
   User-Agent: curl/8.6.0
   Accept: */*

   HTTP/1.1 200 OK
   Content-Type: text/html; charset=utf-8
   Date: Sun, 04 Aug 2024 18:58:50 GMT
   Content-Length: 42
   Connection: close

   <html><body>b991ee98230c58c0</body></html>
   ```
   In the case of HTTP/HTTPS, this is the interaction *request* and *response*.

 :::

