FROM golang:alpine AS builder

WORKDIR /app
COPY src/ /app
RUN go get . && go build -o main main.go

FROM alpine

RUN apk add --no-cache espeak lame
RUN adduser -D tts

COPY --from=builder /app/main /tts

USER tts:tts
WORKDIR /home/tts

ENTRYPOINT [ "/tts" ]
