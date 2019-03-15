#!/usr/bin/env bash
ENV_PASSWORD="${ENV_PASSWORD?Please provide an ENV_PASSWORD}"

docker-compose run --rm gpg \
  --decrypt \
  --batch \
  --passphrase "$ENV_PASSWORD" \
  --out .env \
  .env.gpg
