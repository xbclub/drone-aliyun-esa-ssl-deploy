from golang:1.25-alpine as prebuild
add . /data
workdir /data
run go mod tidy && CGO_ENABLED=0 go build -v -o app -ldflags "-s -w" --trimpath

from alpine:latest
COPY --from=prebuild /data/app /app/
ENTRYPOINT ["/app/app"]