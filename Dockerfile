FROM golang:1.20.8-bullseye
COPY . .
RUN bash build.sh arc-consumer /arc-consumer

FROM golang:1.20.8-bullseye
COPY --from=builder /arc-consumer /arc-consumer
COPY arc-consumer.toml /arc-consumer.toml

COPY pkg/engine/cpp/libs/ /usr/lib/
COPY swagger /swagger
COPY web /web
COPY mod /mod
WORKDIR /
EXPOSE  8972
HEALTHCHECK --interval=30s --timeout=15s \
    CMD curl --fail http://localhost:80/health || exit 1
ENTRYPOINT [ "/arc-consumer" ]
CMD ["run"]
