FROM golang:alpine AS builder

WORKDIR /app
COPY src/ /app
RUN go get . && go build -o main main.go

FROM alpine

RUN apk add --no-cache espeak lame

COPY --from=builder /app/main /tts

ENTRYPOINT [ "/tts" ]
