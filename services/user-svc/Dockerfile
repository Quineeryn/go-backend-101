# --- Builder ---
FROM golang:1.24.6 AS build
WORKDIR /src

# Copy file modul dulu agar cache dependency kepake
COPY go.mod go.sum ./
# Jika kamu juga pakai go.work, uncomment baris berikut:
# COPY go.work ./
RUN go mod download

# Copy seluruh repo (karena kita build dari path ./services/user-svc)
COPY . .

# Build biner statik
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -buildvcs=false -o /out/user-svc ./services/user-svc

RUN GOBIN=/out go install github.com/grpc-ecosystem/grpc-health-probe/@latest
# --- Runner minimal ---
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/user-svc /user-svc
COPY --from=build /out/grpc-health-probe /bin/grpc_health_probe
EXPOSE 50051
ENTRYPOINT ["/user-svc"]
