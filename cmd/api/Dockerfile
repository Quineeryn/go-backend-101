# --- Builder ---
FROM golang:1.24.6 AS build
WORKDIR /src

# Copy file modul dulu agar cache dependency kepake
COPY go.mod go.sum ./
# Jika kamu juga pakai go.work, uncomment baris berikut:
# COPY go.work ./
RUN go mod download

# Copy seluruh repo (karena kita build dari path ./cmd/api)
COPY . .

# Build biner statik
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -buildvcs=false -o /out/api ./cmd/api

# --- Runner minimal ---
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/api /api

EXPOSE 8080
ENTRYPOINT ["/api"]
