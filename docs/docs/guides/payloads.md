---
sidebar_position: 3
---

# Payloads

```
Main commands
  /new         Create a new payload
  /list        List payloads
  /mod         Modify existing payload
  /del         Delete payload
  /clr         Delete multiple payloads
```

- Payloads are just unique domains with associated names.
enable/disable some protocols for payload, add DNS records, etc.) and to distinguish alerts.
- Payload name (`<NAME>`) is used to identify the payload in commands (when you enable/disable some protocols for the payload,
add DNS records, etc.) and to distinguish alerts.
- You can enable/disable alerts for specific protocols for the payload.
information about all alerts for the payload using CLI tool or API.
- You can enable event logging for the payload. In this case all events will be stored in the Sonar server database and you will be able
to get full information about all alerts for the payload using the CLI tool or API.

## Create payload

```
Create a new payload identified by NAME

Usage:
  /new NAME [flags]

Flags:
  -e, --events              Store events in database
  -h, --help                help for new
  -p, --protocols strings   Protocols to notify (default [dns,http,smtp,ftp])
```

### Create new payload

```
/new <NAME>
```

- `<NAME>` — could be any string (use quotes if you want to have spaces in name). It is used to
identify your payload in notifications and other commands.

![Payload creation](../assets/create_payload_telegram_dark.png#gh-dark-mode-only)![Payload creation](../assets/create_payload_telegram_light.png#gh-light-mode-only)

- You will receive notifications about all DNS, HTTP(s), SMTP, FTP interactions containing your
unique subdomain (`d14a68e4` in example).

### Create payload and enable alerts only for selected protocols

```
/new <NAME> -p http,dns
```

![Payload creation with protocols](../assets/create_payload_protocols_telegram_dark.png#gh-dark-mode-only)![Payload creation with protocols](../assets/create_payload_protocols_telegram_light.png#gh-light-mode-only)

### Create payload and enable event logging

```
/new <NAME> -e
```

- By default events are not stored in the database. Event storage is usetul it you want to
automate something and retrieve all events for your subdomain using API or CLI tool.

![Payload creation with events logging](../assets/create_payload_events_telegram_dark.png#gh-dark-mode-only)![Payload creation with events logging](../assets/create_payload_events_telegram_light.png#gh-light-mode-only)


## List payloads

```
List payloads whose NAME contain SUBSTR

Usage:
  /list [SUBSTR] [flags]

Flags:
  -h, --help   help for list
```

### List all payloads

```
/list
```

![List payloads](../assets/list_payloads_telegram_dark.png#gh-dark-mode-only)![List payloads](../assets/list_payloads_telegram_light.png#gh-light-mode-only)

### List payloads containing "SUBSTR" in name

```
/list <SUBSTR>
```

![List payloads](../assets/list_payloads_filter_telegram_dark.png#gh-dark-mode-only)![List payloads](../assets/list_payloads_filter_telegram_light.png#gh-light-mode-only)

## Modify payload

```
Modify existing payload identified by NAME

Usage:
  /mod NAME [flags]

Flags:
  -e, --events              Store events in database
  -h, --help                help for mod
  -n, --name string         Payload name
  -p, --protocols strings   Protocols to notify
```

### Change the protocols for the payload for which you want to be alerted

```
/mod <NAME> -p smtp
```

![Modify payload's protocols](../assets/modify_payload_telegram_dark.png#gh-dark-mode-only)![Modify payload's protocols](../assets/modify_payload_telegram_light.png#gh-light-mode-only)

### Enable events logging for the payload

```
/mod <NAME> -e
```

![Enable events logging for payload](../assets/modify_payload_events_telegram_dark.png#gh-dark-mode-only)![Enable events logging for payload](../assets/modify_payload_events_telegram_light.png#gh-light-mode-only)

## Delete payload

```
Delete payload identified by NAME

Usage:
  /del NAME [flags]

Flags:
  -h, --help   help for del
```

### Delete single payload by name

```
/del <NAME>
```

## Clear payloads

```
Delete payloads that have a SUBSTR in their NAME

Usage:
  /clr [SUBSTR] [flags]

Flags:
  -h, --help   help for clr
```

### Delete all payloads 

```
/clr
```

![Delete all payloads](../assets/clear_all_telegram_dark.png#gh-dark-mode-only)![Delete all payloads](../assets/clear_all_telegram_light.png#gh-light-mode-only)

### Delete payloads containing "SUBSTR" in name

```
/clr <SUBSTR>
```

![Delete payloads containing "SUBSTR"](../assets/clear_telegram_dark.png#gh-dark-mode-only)![Delete payloads containing "SUBSTR"](../assets/clear_telegram_light.png#gh-light-mode-only)
