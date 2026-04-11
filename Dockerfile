FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN mkdir /dist \
	&& cd /app/cmd/flob \
	&& go work init && go work use -r ../ \
	&& GOARCH=amd64 go build -o /dist/amd64 . \
	&& GOARCH=arm64 go build -o /dist/arm64 .

ARG TARGETARCH
RUN "/dist/${TARGETARCH}" version

FROM scratch
COPY --from=builder /dist /
