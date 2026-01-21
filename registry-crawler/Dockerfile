FROM golang:1.25.3 AS go-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o start-app

# Go build complete

#TODO Switch to distroless later
FROM debian:bookworm-slim
WORKDIR /app

COPY --from=go-build /app/start-app ./

#TODO STart debug
RUN apt update
RUN apt install curl -y
#TODO End debug

# Copied binary to distroless image

CMD [ "/app/start-app" ]

EXPOSE 50051