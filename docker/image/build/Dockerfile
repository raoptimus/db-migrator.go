FROM golang:1.24-alpine as build

RUN apk add --update make bash file

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

RUN apk add --no-cache tzdata ca-certificates
ENV TZ UTC

WORKDIR /usr/bin
COPY --from=build /go/src/db-migrator/.build .
