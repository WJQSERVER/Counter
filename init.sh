#!/bin/bash

if [ ! -f /data/caddy/config/Caddyfile ]; then
    cp /data/caddy/Caddyfile /data/caddy/config/Caddyfile
fi

/data/caddy/caddy run --config /data/caddy/config/Caddyfile > /data/caddy/log/caddy-run.log 2>&1 &

/data/counter/counter > /data/counter/log/counter.log 2>&1 &

while [[ true ]]; do
    sleep 1
done    
