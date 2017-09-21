FROM debian:9

RUN apt-get update && apt-get install -y curl netcat

COPY start.sh /opt/app/

ENTRYPOINT ["/opt/app/start.sh"]