# TODO: Use avalanche as submodule instead of cloning in Dockerfile for manual deployments

FROM timbru31/node-alpine-git:gallium AS bundler
WORKDIR /usr/src

RUN git clone https://github.com/ArcticOJ/avalanche --depth 1 --single-branch

WORKDIR /usr/src/avalanche

RUN yarn install --immutable --immutable-cache
RUN yarn build
RUN yarn export -o bundle

FROM golang:alpine AS builder
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=bundler /usr/src/avalanche/bundle ./avalanche/

RUN go build -o ./out/avalanche -ldflags "-s -w" main.go

FROM alpine AS runner
WORKDIR /avalanche

COPY --from=builder /usr/src/app/out/avalanche ./

ENTRYPOINT ["/avalanche/avalanche"]

