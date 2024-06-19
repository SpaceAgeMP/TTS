FROM golang:alpine AS builder

WORKDIR /app
COPY . /app
RUN go get . && go build -o main main.go

FROM alpine

RUN apk add --no-cache espeak lame curl
RUN adduser -D tts

COPY --from=builder /app/main /app/tts

USER tts:tts
WORKDIR /app
VOLUME /app/out
RUN mkdir -p /app/out && chown tts:tts /app/out
ENTRYPOINT [ "/app/tts" ]
