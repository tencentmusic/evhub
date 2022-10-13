FROM alpine:3.15

RUN apk --update add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*

COPY ./bin/processor /bin/

EXPOSE 9000

ENTRYPOINT [ "/bin/processor" ]