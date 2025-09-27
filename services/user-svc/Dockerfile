FROM golang:1.24.6 AS build
WORKDIR /src

# cache deps
COPY go.mod go.sum ./
# COPY go.work ./   # kalau kamu pakai go.work, uncomment
RUN go mod download

# copy source
COPY . .

# build service (static)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -buildvcs=false -o /out/user-svc ./services/user-svc

# build gRPC health probe (static) — pakai go install, tanpa /v2
RUN CGO_ENABLED=0 GOOS=linux GOBIN=/out go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.40

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/user-svc /user-svc
# binary hasil go install bernama "grpc-health-probe"
COPY --from=build /out/grpc-health-probe /grpc_health_probe
EXPOSE 50051
ENTRYPOINT ["/user-svc"]
