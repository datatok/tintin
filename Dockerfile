FROM scratch

ADD ./_dist/linux-amd64/tintin  /usr/local/bin/tintin
ADD VERSION                     /opt/tintin/
ADD templates                   /opt/tintin/

WORKDIR /opt/tintin

EXPOSE 8080/tcp

#HEALTHCHECK --interval=5m --timeout=3s \
#  CMD curl -f http://localhost:8080/status || exit 1

CMD ["tintin", "server"]
