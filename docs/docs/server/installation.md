---
sidebar_position: 1
---

# Installation

Sonar is self-hosted only, so to get started you need to install and configure the server part of it first.

## Requirements

To install the Sonar Server, you must have:

- A Linux server with a public IP address (`<IP>`).
- A registered domain name (`<DOMAIN>`).

## Docker compose

```yml
services:
  sonar:
    restart: always
    image: ghcr.io/nt0xa/sonar:latest
    ports:
      - 21:21       # FTP
      - 25:25       # SMTP
      - 53:53/udp   # DNS
      - 80:80       # HTTP
      - 443:443     # HTTPS
      - 31337:31337 # REST API
    volumes:
      - ./tls:/opt/app/tls               # TLS certificates persistance
      - ./config.yml:/opt/app/config.yml # Config file: see "Configuration"

  db:
    image: postgres:16
    restart: always
    environment:
      POSTGRES_USER: sonar
      POSTGRES_PASSWORD: <POSTGRES_PASSWORD>
      POSTGRES_DB: sonar
    volumes:
      - ./postgres:/var/lib/postgresql/data # DB persistance
```

