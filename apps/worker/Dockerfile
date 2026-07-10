FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -buildvcs=false -o /out/worker ./apps/worker/cmd/worker

FROM alpine:3.22

RUN addgroup -S dentaldesk && adduser -S dentaldesk -G dentaldesk
WORKDIR /app
COPY --from=build /out/worker /app/worker
USER dentaldesk
CMD ["/app/worker"]
