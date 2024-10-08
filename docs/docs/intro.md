---
sidebar_position: 1
---

# Introduction

Sonar is a security researcher's Swiss army knife for finding and exploiting vulnerabilities that require out-of-band interactions.
It is similar to [Burp Collaborator](https://portswigger.net/burp/documentation/collaborator) or [interactsh](https://github.com/projectdiscovery/interactsh), but offers some useful additional features.

## Features

- Ability to create named payloads and receive notifications in the messenger of choice of all interactions with these payloads via DNS, HTTP, FTP and SMTP protocols.
- Currently supported messengers: Telagram, Lark.
- Ability to manage payloads and configure payloads via the messenger of choice or CLI tool.
- Configurable DNS responses with the ability to return multiple records for a name or set up DNS rebinding.
- Configurable HTTP responses: static or dynamic using Go template language.
- Automatic TLS certificates with Let's Encrypt.
- Support for multiple users. Currently there are only two roles: admin and regular user.
- REST API.


