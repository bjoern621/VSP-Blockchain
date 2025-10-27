FROM golang:1.25.3 AS go-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o start-app

# Go build complete

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=go-build /app/start-app ./

# Copied binary to distroless image

CMD [ "/app/start-app" ]

EXPOSE 8080