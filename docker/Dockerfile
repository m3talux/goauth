# =====================
# Target 'build-env'
# =====================
FROM golang:1.21-alpine AS build-env
RUN apk --no-cache add build-base git gcc
ARG GITLAB_TOKEN="${GITLAB_TOKEN}"
RUN git config --global url.https://oauth2:"${GITLAB_TOKEN}"@gitlab.kardinal.ai/.insteadOf https://gitlab.kardinal.ai/

WORKDIR /src

# Copy the go.mod and go.sum to download all the dependencies
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -ldflags="-s -w"  -o goapp

WORKDIR /app
RUN cp /src/goapp .

# =====================
# Target 'prod'
# =====================

FROM alpine as PROD
RUN apk --no-cache add tzdata
WORKDIR /app
ENV PORT 80
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp
EXPOSE 80
