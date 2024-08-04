---
sidebar_position: 2
---

# Notifications

## DNS

- DNS listener runs on port 53.
- You will receive notifications for any DNS queries of your payload's domain (e.g. `d14a68e4.sonar.test`)
  and for any queries of its subdomains (e.g. `test.d14a68e4.sonar.test`).
- DNS interaction notification is a dig-like representation of the question and answer.

  ![DNS notification example](../assets/dns_notification_dark.png#gh-dark-mode-only)![DNS notification example](../assets/dns_notification_light.png#gh-light-mode-only)
