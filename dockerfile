FROM golang:1.26 AS build-stage

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ---------- Runtime stage ----------
FROM alpine:3.20

WORKDIR /app

# copy binary from builder
COPY --from=build-stage /app/server .

EXPOSE 8080

CMD ["./server"]
