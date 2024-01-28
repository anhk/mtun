FROM golang:alpine as builder
ENV WORKPATH /opt/mtun/
ENV CGO_ENABLED=0

RUN apk update && apk add upx make

COPY . $WORKPATH
WORKDIR $WORKPATH
RUN make clean && make
RUN upx mtun

FROM alpine:latest
ENV WORKPATH /opt/mtun/

COPY --from=builder $WORKPATH/mtun /usr/bin/

CMD ["/usr/bin/mtun", "server", ""]