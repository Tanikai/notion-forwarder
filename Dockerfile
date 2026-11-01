FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/notion-forwarder .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/notion-forwarder ./notion-forwarder

EXPOSE 3000

ENTRYPOINT ["./notion-forwarder"]
