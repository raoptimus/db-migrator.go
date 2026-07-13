FROM golang:1.26-alpine
RUN apk add --no-cache  \
    tzdata  \
    ca-certificates  \
    make \
    build-base \
    git \
    bash \
    file \
    jq

RUN export BINDIR=/go/bin  \
    && export PATH=$PATH:$GOPATH/bin

#RUN go install github.com/boumenot/gocover-cobertura@latest  \
#    && go install github.com/jstemmer/go-junit-report@latest

# Install golangci-lint via the Go module proxy (reliable + integrity-checked by GOSUMDB).
# Avoids the flaky/rate-limited GitHub release binary download used by the install script.
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

RUN go env -w GOFLAGS=-buildvcs=false && \
    go env -w CGO_ENABLED=0
