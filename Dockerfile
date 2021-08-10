FROM golang:alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

COPY . .

RUN GOCGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o klottr

FROM scratch

COPY --from=build /build/klottr /

CMD ["/klottr"]

EXPOSE 3000