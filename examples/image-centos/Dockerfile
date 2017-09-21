FROM centos:7

RUN yum install -y curl netcat

COPY start.sh /opt/app/

ENTRYPOINT ["/opt/app/start.sh"]