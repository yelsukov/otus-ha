########################
# STEP 1: build binary #
########################
FROM golang:1.15-alpine AS builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/backend

# Redundant copy for caching vendors download
COPY go.mod go.sum ./
# Fetch dependencies.
RUN go mod download
RUN go mod verify

COPY . .

# Specify version
ARG VERSION
# Specify the backend token
ARG TOKEN

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X github.com/yelsukov/otus-ha/dialogue/vars.VERSION=${VERSION} -X github.com/yelsukov/otus-ha/dialogue/vars.TOKEN=${TOKEN}" -o /go/bin/dialogue

###############################
# STEP 2: build a small image #
###############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy binary file
COPY --from=builder /go/bin/dialogue /go/bin/dialogue

# Use an unprivileged user.
USER ${USER}:${USER}

# Run the binary.
ENTRYPOINT ["/go/bin/dialogue"]
