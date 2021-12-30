# Build a static Go app binary first
FROM golang:1.17 as builder

#RUN apk update && apk add --no-cache git

ENV USER=appuser
ENV UID=1001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/app/
COPY ./chain ./chain
COPY ./config ./config
COPY ./go.* .
COPY main.go .

# fetch dependencies
RUN go mod download
RUN go mod tidy
RUN go mod verify

# strip debug data from the final build
RUN go build -ldflags="-s -w" -o /app/

RUN mkdir -p /app/conf && touch /app/conf/.keep && chown -R appuser:appuser /app

FROM gcr.io/distroless/base

USER appuser:appuser

# import user and group files from above
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /app/tdameritrade-alerter /app/
COPY --from=builder /app/conf/.keep /app/conf/

USER appuser:appuser
VOLUME /app/conf
ENTRYPOINT ["/app/tdameritrade-alerter"]
CMD ["/app/conf"]