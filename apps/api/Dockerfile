FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -buildvcs=false -o /out/api ./apps/api/cmd/api
RUN go build -buildvcs=false -o /out/migrate ./apps/api/cmd/migrate

FROM alpine:3.22

RUN addgroup -S dentaldesk && adduser -S dentaldesk -G dentaldesk
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY infra/migrations /app/infra/migrations
COPY apps/api/render-start.sh /app/render-start.sh
RUN chmod +x /app/render-start.sh
USER dentaldesk
EXPOSE 8080
CMD ["/app/render-start.sh"]
