FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pr-service ./cmd/main.go
EXPOSE 8080
CMD ["./pr-service"]
