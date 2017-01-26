FROM quay.io/prometheus/busybox:latest
MAINTAINER Vincent Van Hollebeke <vincent@compuscene.org>

COPY json_exporter /bin/json_exporter

EXPOSE 8888
ENTRYPOINT [ "/bin/json_exporter" ]
