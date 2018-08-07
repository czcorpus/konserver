# KonServer

KonServer is an asynchronous task queue and calculation status notifier (supporting WebSocket)
for KonText [KonText](https://github.com/czcorpus/kontext). It can be used instead
of [Celery](http://www.celeryproject.org/) along with KonText. Celery support will remain for
future versions of KonText.

:construction:
Please note that this project is under development. For production deployment of KonText
use Celery.
:construction:


## Configuration

### KonText

```xml
<kontext>
    ...
    <calc_backend>
        <type>konserver</type>
        <conf>/var/www/kontext/conf/konserverconfig.py</conf>
        <status_service_url>ws://127.0.0.1:8083/kontext/atn/ws</status_service_url>
        <time_limit>120</time_limit>
    </calc_backend>
    ...
</kontext>
```

### application

Please refer to [config.sample.json](./config.sample.json).

### systemd

```
[Unit]
Description=konserver
After=network.target

[Service]
Type=simple
Restart=on-failure
RestartSec=30
User=www-data
ExecStart=/bin/bash -c '/opt/go/bin/konserver /opt/konserver/config.json'
ExecStop=/bin/kill -s TERM $MAINPID
ExecReload=/bin/kill -s HUP $MAINPID

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
