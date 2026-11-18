---
sidebar_position: 6
---

# Events

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
  -a, --after int        Events with index greater than this, oldest first
  -b, --before int       Events with index less than this, newest first
  -h, --help             help for list
  -l, --limit uint       Limit (default 10)
  -o, --offset uint      Offset
  -p, --payload string   Payload name
```

### Get last 10 events for payload

```
/events list -p <NAME>
```

![List events for payload](../assets/events_list_telegram_dark.png#gh-dark-mode-only)![List events for payload](../assets/events_list_telegram_light.png#gh-light-mode-only)

### Get last N events for payload

```
/events list -p <NAME> -l <N>
```

![List last N events for payload](../assets/events_list_n_telegram_dark.png#gh-dark-mode-only)![List last N events for payload](../assets/events_list_n_telegram_light.png#gh-light-mode-only)

### Get first N events for payload

```
/events list -p <NAME> -l <N> -a 0
```

![List first N events for payload](../assets/events_list_first_telegram_dark.png#gh-dark-mode-only)![List first N events for payload](../assets/events_list_first_telegram_light.png#gh-light-mode-only)

`--limit` is at most 100. To page through events, pass the index of the last event you received as `--after` (oldest first) or `--before` (newest first). `--after` and `--before` cannot be combined with each other or with `--offset`.

### Get N events after Mth event

```
/events list -p <NAME> -l <N> -a <M>
```

![List N events after Mth event for payload](../assets/events_list_mn_telegram_dark.png#gh-dark-mode-only)![List N events after Mth event for payload](../assets/events_list_mn_telegram_light.png#gh-light-mode-only)

### Get N events before Mth event

```
/events list -p <NAME> -l <N> -b <M>
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
