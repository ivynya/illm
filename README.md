# illm

Internet LLM (illm) lets you access your locally run LLM from any computer with a web browser and internet connection, even if you aren't on the same network.

It does this by running a server on a cloud instance that acts as a forwarder between your local machine and the internet. The server is responsible for authenticating clients and forwarding requests to your local machine, and forwarding responses back to the client.

This repository's reference implementation is designed to work with [ollama](https://ollama.ai) as the provider and [ivynya/aura](https://github.com/ivynya/aura) as the user interface.

> ⚠️ This project should mostly be treated as a proof of concept. You may run into stability issues, especially if you share your LLM provider with many people at once.

## Architecture

1. You host an `illm/server` instance on a cloud provider and expose it to the internet on a domain (e.g. `illm.example.com`).
2. You run `illm/client` on your local machine and configure it to your server. The client connects to the server at `/aura/provider`, identifying itself as an LLM provider.
3. You connect to `/aura/client` using an illm client like [Aura](https://github.com/ivynya/aura) and authenticate to the server. Now, requests will be pipelined from the client to the server to the provider and back.
4. Requests from clients are sent as JSON with an `action` and other parameters. See `/internal/types.go`. Requests are tagged by the server with a unique ID (Tag) corresponding to each client connection, then sent to the provider. The provider is responsible for processing the request and sending back a Request object with the same Tag. The server then sends the response back to the client with a matching Tag.

Because the server hosts websocket endpoints, connections can be made from anywhere without reverse proxying.

## Usage

This repository contains a reference implementation of an illm provider (in `/client`). It needs ollama installed on your local machine running at localhost:11434 and will make API requests outside of the docker container to that URL. It is designed to work with the reference implementation of the user client, [Aura](https://github.com/ivynya/aura).

Example docker compose file for running the server on your cloud instance:

```yaml
version: "3.8"

services:
  illm:
    image: ghcr.io/ivynya/illm/server:latest
    ports:
      - 8080:3000
    restart: unless-stopped
    environment:
      - USERNAME=admin
      - PASSWORD=password
```

Example docker compose file for running the client on your local machine:

```yaml
version: "3.8"

services:
  illm:
    image: ghcr.io/ivynya/illm/client:latest
    restart: unless-stopped
    environment:
      - AUTH=<a base64 encoded username:password>
      - IDENTIFIER=your-computer-name
      - ILLM_SCHEME=<ws|wss>
      - ILLM_HOST=illm.example.com
      - ILLM_PATH=/aura/provider
```

Run the server first, then the client. The client should log that it is connected. Then, if you don't want to write your own client, set up [Aura](https://github.com/ivynya/aura) as described in the README.

## Development

This repository uses a modified subset of [langchaingo](https://github.com/tmc/langchaingo)'s ollama implementation in the reference client. It was modified to return additional data during generation, since the original returns text only (without extra info like tokens, duration, and context). It was also modified to accept chat context as a parameter.
