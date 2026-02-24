# What is Hookbridge?
Hookbridge is a tunneling service (like ngrok, or cloudflare), except with a twist: it allows you to connect multiple clients to the same public IP. When an Http request is sent to the public IP (hookbridge),the Http request is forwarded to every single client that is connected to that tunnel. One client is chosen as the primary responder - that client will return the Http response, which gets sent back to the initial sender.

# Dockerfile.tunnel
Run from project root:
```docker build --file Dockerfile.tunnel -t tunnel-image .```

Make sure to run this command before starting the server, since the server automatically spins up this docker container for each new tunnel.

# Usage
First, make sure the server is running by default. The server is responsible for spinning up tunnels and forwarding Http Requests to the tunnel (which then forwards to clients). By default, the server runs on port 8080.
```bash
go run cmd/server
```

Then, to connect to a tunnel, run:
```bash
hookbridge connect --name <tunnel-name> --port <PORT_NUM>
```

Afterwards, all HttpRequests sent to `<hookbridge-server-ip>/tunnel/<tunnel-name>/*` will be forwarded to `localhost:<specified port number>`

Then, one connected client will be chosen to send back an HTTPResponse
