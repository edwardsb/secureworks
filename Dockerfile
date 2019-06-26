FROM golang:1.12-stretch as build

RUN mkdir /usr/local/share/GeoIP
RUN go get -u github.com/maxmind/geoipupdate/cmd/geoipupdate
RUN go get -u github.com/go-bindata/go-bindata/...

COPY GeoIP.conf /usr/local/etc

RUN geoipupdate

WORKDIR /src
COPY . .
RUN make build

FROM alpine:3.10
# https ssl certs, curl for healthcheck
RUN apk --update upgrade && apk --no-cache add curl && apk --no-cache add ca-certificates
RUN mkdir /usr/local/share/GeoIP
RUN mkdir /opt/app
RUN mkdir /var/lib/data  # sqlite file
VOLUME /var/lib/data
EXPOSE 3000
# we need CGO for sqlite, but musl is in alpine, not gclib, so we need to symlink
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=build /usr/local/share/GeoIP /usr/local/share/GeoIP
COPY --from=build /src/secureworks /opt/app
CMD /opt/app/secureworks server