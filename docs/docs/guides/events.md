---
sidebar_position: 6
---

# Events

:::warning

You can only view events for payloads with events logging enabled, see: [Enable events](/sonar/guides/payloads#create-payload-and-enable-event-logging)

:::

```
View events

Usage:
 /events [command]

Available Commands:
  /list        List payload events
  /get         Get payload event by INDEX

Flags:
  -h, --help   help for events

Use "/events [command] --help" for more information about a command.
```

## List events for payload

```
List payload events

Usage:
  /events list [flags]

Flags:
  -a, --after int        After ID
  -b, --before int       Before ID
  -c, --count uint       Count of events (default 10)
  -h, --help             help for list
  -p, --payload string   Payload name
  -r, --reverse          List events in reversed order
```

### Get last 10 events for payload

```
/events list -p <NAME>
```

![List events for payload](../assets/events_list_telegram_dark.png#gh-dark-mode-only)![List events for payload](../assets/events_list_telegram_light.png#gh-light-mode-only)

### Get last N events for payload

```
/events list -p <NAME> -c <N>
```

![List last N events for payload](../assets/events_list_n_telegram_dark.png#gh-dark-mode-only)![List last N events for payload](../assets/events_list_n_telegram_light.png#gh-light-mode-only)

### Get first N events for payload

```
/events list -p <NAME> -c <N> -r
```

![List first N events for payload](../assets/events_list_first_telegram_dark.png#gh-dark-mode-only)![List first N events for payload](../assets/events_list_first_telegram_light.png#gh-light-mode-only)

### Get N events after Mth event

```
/events list -p <NAME> -c <N> -a <M>
```

![List N events after Mth event for payload](../assets/events_list_mn_telegram_dark.png#gh-dark-mode-only)![List N events after Mth event for payload](../assets/events_list_mn_telegram_light.png#gh-light-mode-only)

### Get N events before Mth event

```
/events list -p <NAME> -c <N> -b <M>
```

![List N events before Mth event for payload](../assets/events_list_mn2_telegram_dark.png#gh-dark-mode-only)![List N events before Mth event for payload](../assets/events_list_mn2_telegram_light.png#gh-light-mode-only)

## Get event

```
Get payload event by INDEX

Usage:
  /events get INDEX [flags]

Flags:
  -h, --help             help for get
  -p, --payload string   Payload name
```

### Get event for payload by index

```
/events get -p <NAME> <INDEX>
```

![Get event for payload by index](../assets/events_get_telegram_dark.png#gh-dark-mode-only)![Get event for payload by index](../assets/events_get_telegram_light.png#gh-light-mode-only)
