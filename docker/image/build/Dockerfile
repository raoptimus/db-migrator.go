FROM golang:1.16-alpine3.12 as build

RUN apk add --update make git bash file

WORKDIR /go/src/db-migrator
COPY . ./

RUN find .
#RUN go mod vendor

ARG VERSION
ARG GIT_TAG
ARG GIT_BRANCH
ARG GIT_COMMIT

RUN make build

FROM alpine:latest AS db-migrator

WORKDIR /usr/bin
COPY --from=build /go/src/db-migrator/.build .
