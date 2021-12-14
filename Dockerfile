FROM alpine:3.11

ADD ./_dist/linux-amd64/tintin /usr/local/bin/tintin
ADD VERSION /opt/tintin/
ADD templates /opt/tintin/

RUN apk update && \
    apk add git 

WORKDIR /opt/tintin

CMD ["tintin", "server"]
