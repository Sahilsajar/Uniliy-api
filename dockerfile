# ---------- DEV STAGE ----------
FROM golang:1.26 AS dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


# ---------- BUILD STAGE ----------
FROM golang:1.26 AS build-stage

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server


# ---------- RUNTIME STAGE ----------
FROM alpine:3.20

WORKDIR /app

COPY --from=build-stage /app/server .

EXPOSE 8080

CMD ["./server"]