FROM golang:1.24.2-alpine AS build

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY backend ./backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cms-api ./backend

FROM alpine:3.21
WORKDIR /app
COPY --from=build /out/cms-api /app/cms-api
EXPOSE 8888
ENTRYPOINT ["/app/cms-api", "-f", "etc/cms-api.yaml"]
