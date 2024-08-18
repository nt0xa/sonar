---
sidebar_position: 2
---

# Alerts

## DNS

- DNS listener runs on port **53**.
- You will receive notifications for any DNS queries of your payload's domain (e.g. `d14a68e4.sonar.test`)
  and for any queries of its subdomains (e.g. `test.d14a68e4.sonar.test`).
- A notification is a dig-like representation of a DNS question and its answer.

  ![DNS notification example](../assets/dns_notification_dark.png#gh-dark-mode-only)![DNS notification example](../assets/dns_notification_light.png#gh-light-mode-only)

## HTTP

- HTTP listener runs on port **80**, HTTPS on port **443**.
- You will receive notifications for any HTTP(s) request that contains your subdomain (e.g. `d14a68e4`) **anywhere in the request**.
  It doesn't matter if it is in the `Host` header or any other header or body.
- ⚠️ HTTP/2 and HTTP/3 are not supported yet, **only HTTP/1.1**.
- ⚠️ You can also use subdomains on your payload domain (i.e. `test.d14a68e4.sonar.test`), but
  the Sonar server won't automatically request a TLS certificate for them (it only requests `*.<DOMAIN>` certificate),
  so remote HTTPS client will most likely get a certificate validation error for them.
- A notification is an HTTP request and its response:

  ![HTTP notification](../assets/http_notification_dark.png#gh-dark-mode-only)![HTTP notification](../assets/http_notification_light.png#gh-light-mode-only)

## SMTP

- SMTP listener runs on port **25**.
- `STARTTLS` is supported.
- You will receive notifications for any SMTP session that contains you subdomain (e.g. `d14a68e4`) **anywhere** in it.
- A notification is a full log of SMTP session:

  ![SMTP notification](../assets/smtp_notification_dark.png#gh-dark-mode-only)![SMTP notification](../assets/smtp_notification_light.png#gh-light-mode-only)

- Additionally you will receive `.eml` file with a content of email. This file can be opened in any email client to view the rendered content.

  ![SMTP notification EML](../assets/smtp_notification_eml_dark.png#gh-dark-mode-only)![SMTP notification EML](../assets/smtp_notification_eml_light.png#gh-light-mode-only)

## FTP

- FTP listener runs on port **21**.
- ⚠️ To receive an FTP notification, your unique domain (e.g. `d14a68e4`) must appear somewhere in the FTP communication log.
  You can achieve this by adding the domain to the user/password or file name, like this: `ftp://d14a68e4:pass@sonar.test` or `ftp://sonar.test/d14a68e4`.
- A notification is a full FTP session log.

  ![FTP notification](../assets/ftp_notification_dark.png#gh-dark-mode-only)![FTP notification](../assets/ftp_notification_light.png#gh-light-mode-only)
