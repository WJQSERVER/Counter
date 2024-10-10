#!/bin/sh

# 检查并复制 Caddyfile
if [ ! -f /data/caddy/config/Caddyfile ]; then
    cp /data/caddy/Caddyfile /data/caddy/config/Caddyfile
fi

# 检查并复制 config.yaml
if [ ! -f /data/counter/config/config.yaml ]; then
    cp /data/counter/config.d/config.yaml /data/counter/config/config.yaml
fi

# 启动 Caddy
/data/caddy/caddy run --config /data/caddy/config/Caddyfile > /data/caddy/log/caddy-run.log 2>&1 &

# 启动 counter 应用
/data/counter/counter > /data/counter/log/counter.log 2>&1 &

# 保持脚本运行
while true; do
    sleep 1
done