FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY . .

RUN apk add --no-cache upx

RUN go mod download

RUN set CGO_ENABLED=0 && go build -ldflags="-s -w" -o bin .

RUN upx --best --lzma bin

FROM scratch

WORKDIR /app

COPY --from=builder /build/bin ./bin

COPY . .

ENTRYPOINT ["/app/bin"]