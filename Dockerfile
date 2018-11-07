FROM  quay.io/prometheus/busybox:latest
LABEL maintainer="The KubeVault Authors <hello@appscode.com>"

COPY vault_exporter /bin/vault_exporter

RUN mkdir -p /conf
COPY vault-mapping /conf/vault-mapping

EXPOSE      9102 9125 9125/udp
ENTRYPOINT  [ "/bin/vault_exporter", "--statsd.mapping-config=/conf/vault-mapping" ]
