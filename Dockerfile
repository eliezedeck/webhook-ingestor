FROM golang:1.19-alpine AS build

ENV CGO_ENABLED=0

COPY . /project/src

RUN \
  cd /project/src && \
  go build -o /webhook-ingestor .

FROM alpine:latest AS final
WORKDIR /
COPY --from=build /webhook-ingestor .
ENTRYPOINT ["/webhook-ingestor"]
