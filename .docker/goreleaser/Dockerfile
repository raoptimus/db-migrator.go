# Runtime image used by GoReleaser (see `dockers` in .goreleaser.yaml).
# GoReleaser places the prebuilt, statically-linked `db-migrator` binary into the
# build context, so this image just wraps it in a minimal Alpine runtime — no
# compilation happens here. Built per-arch and combined into a multi-arch manifest.
FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=UTC
COPY db-migrator /usr/bin/db-migrator
ENTRYPOINT ["/usr/bin/db-migrator"]
