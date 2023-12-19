ARG GO_IMAGE_VERSION
FROM golang:${GO_IMAGE_VERSION}-alpine
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

ARG GO_LINT_VERSION
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v${GO_LINT_VERSION}

RUN go env -w GOFLAGS=-buildvcs=false && \
    go env -w CGO_ENABLED=0
