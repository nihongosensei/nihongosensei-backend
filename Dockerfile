FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

ENV PORT="3000"

WORKDIR /app

COPY --chmod=555 app ./

USER 65532:65532

EXPOSE 3000

ENTRYPOINT ["/app/app"]
