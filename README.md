# Dockerfile.tunnel
Run from project root:
```docker build --file Dockerfile.tunnel -t tunnel-image .```


# Usage
Run:
```bash
hookbridge connect --name <tunnel-name> --port <PORT_NUM>
```

Afterwards, all HttpRequests sent to `<hookbridge-server-ip>/tunnel/<tunnel-name>/*` will be forwarded to `localhost:<specified port number>`

Then, one connected client will be chosen to send back an HTTPResponse