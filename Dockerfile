# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23.2 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /optiguide

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /optiguide /optiguide
COPY --from=build-stage /app/guide.xlsx /guide.xlsx
COPY --from=build-stage /app/static /static
COPY --from=build-stage /app/templates /templates

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/optiguide"]
