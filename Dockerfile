FROM golang:1.25.5 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/todo-scheduler ./main.go

FROM alpine:latest

WORKDIR /app

RUN mkdir -p /app/data

ENV TODO_DBFILE=/app/data/scheduler.db

COPY --from=builder /out/todo-scheduler /app/todo-scheduler
COPY web /app/web

CMD ["/app/todo-scheduler"]
