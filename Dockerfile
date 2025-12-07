FROM alpine:3.23

ENV PORT="3000"

WORKDIR /app

RUN adduser -D -u 1000 go

USER go

COPY --chown=go:go app ./

RUN chmod +x ./app

EXPOSE 3000

ENTRYPOINT ["./app"]
