FROM golang:1.14.3-alpine AS build
WORKDIR /src
COPY ./ ./
RUN GOOS=linux GOARCH=amd64 go build -o sim ./cmd/sim/sim.go
EXPOSE 8089
CMD ["./sim"]
