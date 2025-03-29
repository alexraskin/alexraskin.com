FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT
ARG BUILD_TIME

RUN --mount=type=cache,id=alexraskin.com-go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=alexraskin.com-go-pkg,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-X 'main.version=$VERSION' -X 'main.commit=$COMMIT' -X 'main.buildTime=$BUILD_TIME'" -o alexraskin.com github.com/alexraskin/alexraskin.com

FROM alpine

RUN apk --no-cache add ca-certificates

COPY --from=build /build/alexraskin.com /bin/alexraskin.com

EXPOSE 8000

CMD ["/bin/alexraskin.com", "-port", "8000"]