# kontext-atn

Kontext-atn (= synchronous task notifier) is an optional WebSocket service for
[KonText](https://github.com/czcorpus/kontext). It can reduce load
on KonText's web workers by handling realtime notifications about background tasks
(currently only concordance calculation).


## Configuration

### application

Please refer to [config.sample.json](./config.sample.json).

### systemd

```
[Unit]
Description=kontext-atn
After=network.target

[Service]
Type=simple
Restart=on-failure
RestartSec=30
User=www-data
ExecStart=/bin/bash -c '/opt/go/bin/kontext-atn /opt/kontext-atn/config.json'
ExecStop=/bin/kill -s TERM $MAINPID

[Install]
WantedBy=multi-user.target
```

### Nginx as a proxy

```
upstream atn_server {
    server localhost:8085 fail_timeout=3;
}

server {
    listen 80;

    # kontext configuration goes here

    location /atn/ {
        proxy_pass http://atn_server;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
    }
}
```

For more info please refer to [NGINX as a WebSocket Proxy](https://www.nginx.com/blog/websocket-nginx/)
