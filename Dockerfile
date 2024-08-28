FROM wjqserver/caddy:latest

RUN mkdir -p /data/www
RUN mkdir -p /data/counter/config
RUN mkdir -p /data/counter/counter
RUN mkdir -p /data/counter/log
RUN wget -O /data/caddy/Caddyfile https://raw.githubusercontent.com/WJQSERVER/Counter/main/Caddyfile
RUN VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER/Counter/main/VERSION) && \
    wget -O /data/counter/counter https://github.com/WJQSERVER/counter/releases/download/$VERSION/counter
RUN wget -O /data/counter/config/config.yaml https://raw.githubusercontent.com/WJQSERVER/Counter/main/config.yaml
RUN wget -O /usr/local/bin/init.sh https://raw.githubusercontent.com/WJQSERVER/Counter/main/init.sh
RUN chmod +x /data/counter/counter
RUN chmod +x /usr/local/bin/init.sh

CMD ["/usr/local/bin/init.sh"]

