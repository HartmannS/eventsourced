FROM eventsourced/builder AS builder

WORKDIR /tmp/eventsourced
COPY . .

ENV GO111MODULE="on"
RUN 2>&1 make test-no-race && make build && upx -q eventsourced

FROM alpine:3.9
COPY --from=builder /tmp/eventsourced/eventsourced /bin/
COPY --from=builder /tmp/eventsourced/config.yml /etc/eventsourced/

RUN addgroup -S worker && adduser -S worker -G worker
USER worker

EXPOSE 2069/tcp
ENTRYPOINT ["/bin/eventsourced"]
