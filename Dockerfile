FROM golang:alpine AS build

WORKDIR /build

COPY . .

RUN GOCGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o /build/klottr
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

FROM scratch

COPY --from=build /build/klottr /klottr
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc_passwd /etc/passwd

USER nobody

CMD ["/klottr"]