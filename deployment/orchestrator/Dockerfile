FROM debian:stretch

EXPOSE 3000

RUN apt update && apt -y install wget curl jq
RUN wget https://github.com/github/orchestrator/releases/download/v3.2.3/orchestrator_3.2.3_amd64.deb \
    && dpkg -i orchestrator_3.2.3_amd64.deb  \
    && mkdir /var/lib/orchestrator && rm -rf /var/lib/apt/lists/*

COPY orc.conf.json /etc/orchestrator.conf.json

WORKDIR /usr/local/orchestrator

COPY setupOrc.sh setupOrc.sh
RUN chmod +x setupOrc.sh

CMD ./orchestrator --config=/etc/orchestrator.conf.json http