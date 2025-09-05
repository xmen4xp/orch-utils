
FROM --platform=linux/amd64 golang:1.24 as builder

ARG API_GW_COMPONENT_NAME 
WORKDIR /app
COPY . /app/
RUN ls -l /app
WORKDIR /app/api-gw
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o bin/${API_GW_COMPONENT_NAME} ./cmd/api-gw

#FROM gcr.io/distroless/static
FROM golang:1.24
WORKDIR /bin
COPY --from=builder /app/api-gw/bin/api-gw .
USER 65532:65532
ENTRYPOINT ["/bin/api-gw"]
