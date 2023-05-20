FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add bash && \
    rm -rf /var/cache/apk*

ADD https://github.com/pressly/goose/releases/download/v3.11.2/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

CMD ["./goose"]