# Replicache TODO Sample

This repository contains a complete Relicache sample that implements a basic mobile Todo app.

## Server

The server is in the `serve` directory. It's a Zeit Now app in Go.

Persistence is against AWS Aurora (MySQL flavor).

### Development

1. Get the Rocicorp AWS credentials and put them in (.aws/credentials) on your machine
2. Add `.env` file to the root of this repository containing:

```
REPLICANT_AWS_ACCESS_KEY_ID=<access key from .aws/credentials>
REPLICANT_AWS_SECRET_ACCESS_KEY=<secret access key from .aws/credentials>
REPLICANT_SAMPLE_TODO_ENV=dev_<your Rocicorp username>
```

3. Run `now dev`

## Deploy

Just commit to origin/master, it is auto-deployed.

Alternately, you can deploy to your own staging environment with:

```
now deploy
```

## Schema

The schema we run against is managed in `schema.go`. Whenever it is changed, the db is dropped and re-created.

We don't currently attempt to migrate data between versions.

# Client

TODO
