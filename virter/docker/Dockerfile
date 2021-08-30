FROM alpine

RUN apk update && \
    apk add --no-cache openssh bash python3

COPY entry.sh gatewaytest.py /
COPY tests /tests

WORKDIR /

ENTRYPOINT /entry.sh
