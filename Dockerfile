FROM golang:1.22.0 as builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -mod=vendor -o /bin/app cmd/main.go

FROM gcr.io/distroless/base as news-api

COPY --from=builder /bin/app /bin/app

EXPOSE 8080

ENTRYPOINT ["/bin/app"]