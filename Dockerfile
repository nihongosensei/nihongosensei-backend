FROM alpine:3.21

ENV PORT="3000"

WORKDIR /app

COPY app ./

RUN chmod +x ./app

EXPOSE 3000

ENTRYPOINT ["./app"]
