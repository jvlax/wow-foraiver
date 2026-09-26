# The endpoint image: a static Go binary and nothing else.
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo docker)" -o /wow-foraiver .

FROM scratch
COPY --from=build /wow-foraiver /wow-foraiver
EXPOSE 8080
ENTRYPOINT ["/wow-foraiver", "serve", "--http", ":8080"]
