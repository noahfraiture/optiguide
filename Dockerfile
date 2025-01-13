FROM golang:1.23.2 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Development stage with air for hot reload
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
CMD ["air", "-c", ".air.toml"]

FROM node:23-alpine AS css-build-stage
WORKDIR /app
RUN npm install -g tailwindcss
COPY tailwind.config.js .
COPY templates ./templates
COPY static ./static
RUN tailwindcss -i ./static/css/input.css -o ./static/css/output.css

FROM base AS build-stage
WORKDIR /go/src/optiguide
COPY main.go go.mod go.sum ./
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /optiguide

# Final stage
FROM alpine:3.21 AS prod
COPY templates /templates
COPY migrations/ /migrations
COPY static/favicon.png /static/favicon.png
COPY static/images /static/images
COPY guide.xlsx /guide.xlsx
COPY --from=build-stage /optiguide /optiguide
COPY --from=css-build-stage /app/static/css/output.css /static/css/output.css

EXPOSE 8080

# USER nonroot:nonroot

ENTRYPOINT ["/optiguide"]

# Choose the target based on the environment
FROM ${TARGET:-prod} AS final
