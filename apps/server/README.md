<p align="center">
  <img src="https://raw.githubusercontent.com/BananoCoin/boompow-next/master/logo.svg" width="300">
</p>

# Server

This is the BoomPoW server, it coordinates work requests from users/services to "providers" that use the client.

## About

The server is written in [GOLang](https://go.dev) and requires Postgres and Redis, you can reference the [docker-compose.yaml](https://github.com/bananocoin/boompow/blob/master/docker-compose.yaml) in the workspace root for details on how to run the server.

It provides a GraphQL API at `/graphql` and the schema can be seen [here](https://github.com/bananocoin/boompow/blob/master/apps/server/graph/schema.graphqls)

A secure websocket endpoint is also available at `/ws/worker` this is the channel that the providers and server use to communicate work requests and responses.

Users are broken up into 2 categories:

1. PROVIDER

Providers are the users who are providing work to BoomPoW in exchange for rewards.

2. REQUESTER

The requesters are work consumers, services that have access to request work from BoomPoW and by proxy the providers.

Work is requested using the `workGenerate` mutation and requires authentication using a service token (not the JWT token returned from the `login` mutation). These tokens can be obtained using the `generateServiceToken` mutation.

There are some layers on protection to prevent users from requesting work.

1. Email must be verified
2. `can_request_work` must be set to true in the database

The second part is intended to happen manually, after a new service requests a key they will be manually approved, after which they can invoke the `generateServiceToken` mutation.
