# syntax=docker/dockerfile:1

# Base image for building CSS files
FROM node:14 AS css-build-stage

WORKDIR /app

RUN npm install -g tailwindcss

COPY . .

RUN tailwindcss -i ./static/css/input.css -o ./static/css/output.css

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
COPY --from=build-stage /app/templates /templates
COPY --from=build-stage /app/migrations/ /migrations
COPY --from=build-stage /app/static/favicon.png /static/favicon.png
COPY --from=build-stage /app/static/images /static/images
COPY --from=css-build-stage /app/static/css/output.css /static/css/output.css

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/optiguide"]
