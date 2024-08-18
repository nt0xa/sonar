---
sidebar_position: 4
---

# DNS records

```
Manage DNS records

Usage:
 /dns [command]

Available Commands:
  /new         Create new DNS records
  /del         Delete DNS record
  /list        List DNS records
  /clr         Delete multiple DNS records

Flags:
  -h, --help   help for dns

Use "/dns [command] --help" for more information about a command.
```

- You can manage DNS records for subdomains of your payload domains.
- There are several record types supported: "A", "АААА", "МХ", "ТХТ", "CNAME", "NS", "CAA".
- You can set TTL for records.
- Wildcard records are supported.
- Multiple records are supported with different strategies:
  - `all` — All values for the record are returned at once.
  - `round-robin` - Values for the record are rotated (first returned, then second, then
third, etc.)
  - `rebind` - Value for the record depends on time between requests. If time > 3s, the first
value is returned, otherwise the second value.

## Create new DNS record

```
Create new DNS records

Usage:
  /dns new VALUES... [flags]

Flags:
  -h, --help              help for new
  -n, --name string       Subdomain
  -p, --payload string    Payload name
  -s, --strategy string   Strategy for multiple records (one of "all", "round-robin", "rebind") (default "all")
  -l, --ttl int           Record TTL (in seconds) (default 60)
  -t, --type string       Record type (one of "A", "AAAA", "MX", "TXT", "CNAME", "NS", "CAA") (default "A")
```

### Create A-record with IP 127.0.0.1 for payload

```
/dns new --payload <NAME> --name <SUBDOMAIN> --type A 127.0.0.1
```

![Create new DNS record](../assets/dns_new_telegram_dark.png#gh-dark-mode-only)![Create new DNS record](../assets/dns_new_telegram_light.png#gh-light-mode-only)

Now `abc.d14a68e4.sonar.test` will respond with IP-address `127.0.0.1` for A-query:

![Test DNS](../assets/dns_test_dark.png#gh-dark-mode-only)![Test DNS](../assets/dns_test_light.png#gh-light-mode-only)

And you will also receive an alert:

![DNS test alert](../assets/dns_test_alert_telegram_dark.png#gh-dark-mode-only)![DNS test alert](../assets/dns_test_alert_telegram_light.png#gh-light-mode-only)

### Create multiple A-records for payload

```
/dns new -p <NAME> -n <SUBDOMAIN> -t A 1.1.1.1 2.2.2.2 3.3.3.3
```

![Create new DNS record with multiple IPs](../assets/dns_new_multiple_telegram_dark.png#gh-dark-mode-only)![Create new DNS record with multiple IPs](../assets/dns_new_multiple_telegram_light.png#gh-light-mode-only)

Now `multiple.d14a68e4.sonar.test` will return all 3 IPs for A record:

![Test DNS multiple records](../assets/dns_test_multiple_dark.png#gh-dark-mode-only)![Test DNS multiple records](../assets/dns_test_multiple_light.png#gh-light-mode-only)

And you will also receive an alert:

![DNS test alert multiple](../assets/dns_test_alert_multiple_telegram_dark.png#gh-dark-mode-only)![DNS test alert multiple](../assets/dns_test_alert_multiple_telegram_light.png#gh-light-mode-only)

### Create wildcard AAAA-record for payload

```
/dns new -p <NAME> -n "*" -t AAAA 2606:2800:220:1:248:1893:25c8:1946
```

![Create new DNS wildcard record](../assets/dns_new_wildcard_telegram_dark.png#gh-dark-mode-only)![Create new DNS wildcard record](../assets/dns_new_wildcard_telegram_light.png#gh-light-mode-only)

Now any query for AAAA record on `*.d14a68e4.sonar.test` will return an IP `2606:2800:220:1:248:1893:25c8:1946`:

![Test DNS wildcard records](../assets/dns_test_wildcard_dark.png#gh-dark-mode-only)![Test DNS wildcard records](../assets/dns_test_wildcard_light.png#gh-light-mode-only)

And you will also receive an alert:

![DNS test alert wildcard](../assets/dns_test_alert_wildcard_telegram_dark.png#gh-dark-mode-only)![DNS test alert wildcard](../assets/dns_test_alert_wildcard_telegram_light.png#gh-light-mode-only)

### Create rebinding record for payload

```
/dns new -p <NAME> -n <SUBDOMAIN> -l 0 -t A -s rebind 1.1.1.1 127.0.0.1
```

- ⚠️ In this case you must set TTL to 0 (`-l 0` or `--ttl 0`) otherwise it won't work.
- `-s` is shorthand for `--strategy`, the default value is `all`, which means "return all
values for this query at once". In this case we use `rebind`, which means "return the first
value (1.1.1.1) if the record hasn't been requested in the last 3 seconds, otherwise return the
next value (127.0.0.1)".
- This can be used to bypass SSRF checks using TOCTOU issues.

![Create new DNS rebind record](../assets/dns_new_rebind_telegram_dark.png#gh-dark-mode-only)![Create new DNS rebind record](../assets/dns_new_rebind_telegram_light.png#gh-light-mode-only)

Here is the result of requesting `rebind.d14a68e4.sonar.test` with delay < 3 seconds between requests:

![Test DNS rebind records](../assets/dns_test_rebind_1_dark.png#gh-dark-mode-only)![Test DNS rebind records](../assets/dns_test_rebind_1_light.png#gh-light-mode-only)

And you will also receive an alert:

![DNS test alert rebind](../assets/dns_test_alert_rebind_telegram_dark.png#gh-dark-mode-only)![DNS test alert rebind](../assets/dns_test_alert_rebind_telegram_light.png#gh-light-mode-only)

## List records

```
List DNS records

Usage:
  /dns list [flags]

Flags:
  -h, --help             help for list
  -p, --payload string   Payload name
```

### List DNS records for payload

```
/dns list -p <NAME>
```

- Every DNS record has an index, which can be used in `/del` command to remove the record.

![List DNS records for payload](../assets/dns_list_telegram_dark.png#gh-dark-mode-only)![List DNS records for payload](../assets/dns_list_telegram_light.png#gh-light-mode-only)

## Delete

```
Delete DNS record identified by INDEX

Usage:
  /dns del INDEX [flags]

Flags:
  -h, --help             help for del
  -p, --payload string   Payload name
```

### Delete DNS record for payload by index

```
/del -p <NAME> <INDEX>
```

![Delete DNS record for payload by index](../assets/dns_del_telegram_dark.png#gh-dark-mode-only)![Delete DNS record for payload by index](../assets/dns_del_telegram_light.png#gh-light-mode-only)

## Clear DNS records for payload

```
/dns clr -p <NAME>
```

![Delete all DNS records for payload](../assets/dns_clear_telegram_dark.png#gh-dark-mode-only)![Delete all DNS records for payload](../assets/dns_clear_telegram_light.png#gh-light-mode-only)
