FROM golang:1.23-alpine
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

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.62.2

RUN go env -w GOFLAGS=-buildvcs=false && \
    go env -w CGO_ENABLED=0
