FROM golang:alpine AS builder

WORKDIR /app
COPY . /app
RUN go get . && go build -trimpath -ldflags '-w -s' -o tts .

FROM alpine

RUN apk add --no-cache espeak lame curl
RUN adduser -D tts

COPY --from=builder /app/tts /app/tts

WORKDIR /app
VOLUME /app/out

RUN mkdir -p /app/out && chown tts:tts /app/out

ENV OUT_DIR=/app/out
ENV LISTEN_ADDR=:4001
USER tts:tts
ENTRYPOINT [ "/app/tts" ]
