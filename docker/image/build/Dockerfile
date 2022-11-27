FROM golang:1.19-alpine3.16 as build

RUN apk add --update make git bash file

WORKDIR /go/src/db-migrator
COPY . ./

RUN find .
RUN go mod tidy

ARG VERSION
ARG GIT_TAG
ARG GIT_BRANCH
ARG GIT_COMMIT

RUN make build

FROM alpine:latest AS db-migrator

WORKDIR /usr/bin
COPY --from=build /go/src/db-migrator/.build .
