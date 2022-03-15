FROM golang:1.17.8-alpine AS builder
RUN apk --no-cache add curl make protoc unzip
ADD . /go/src/delegatedidentity
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
WORKDIR /go/src/delegatedidentity
RUN make

FROM alpine AS callee
RUN apk --no-cache add dumb-init
COPY --from=builder /go/src/delegatedidentity/build/bin/callee /opt/delegatedidentitypoc/bin/callee
COPY --from=builder /go/src/delegatedidentity/configs/callee/* /etc/callee/
WORKDIR /opt/delegatedidentitypoc
ENV CONFIG_DIR=/etc/callee
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/opt/delegatedidentitypoc/bin/callee"]

FROM alpine AS caller
RUN apk --no-cache add dumb-init
COPY --from=builder /go/src/delegatedidentity/build/bin/caller /opt/delegatedidentitypoc/bin/caller
WORKDIR /opt/delegatedidentitypoc
ENV CONFIG_DIR=/etc/caller
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/opt/delegatedidentitypoc/bin/caller"]

FROM alpine AS proxy
RUN apk --no-cache add dumb-init
COPY --from=builder /go/src/delegatedidentity/build/bin/proxy /opt/delegatedidentitypoc/bin/proxy
COPY --from=builder /go/src/delegatedidentity/configs/proxy/* /etc/proxy/
RUN mkdir -p /run/proxy
WORKDIR /opt/delegatedidentitypoc
ENV CONFIG_DIR=/etc/proxy
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/opt/delegatedidentitypoc/bin/proxy"]

FROM gcr.io/spiffe-io/spire-server:1.1.3 AS spire-registrant
ADD ./scripts/create-spire-entries.sh /scripts/create-spire-entries.sh
ENTRYPOINT ["/scripts/create-spire-entries.sh"]
