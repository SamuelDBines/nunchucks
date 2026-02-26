FROM golang:1.24-alpine AS go-builder
WORKDIR /src
COPY go ./go
RUN cd go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/nunchucks ./cmd/nunchucks

FROM node:20-alpine
WORKDIR /app

ENV NODE_ENV=production
ENV APP_PORT=3000
ENV NUNCHUCKS_BIN=/usr/local/bin/nunchucks

COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --prod --frozen-lockfile --ignore-scripts

COPY app ./app
COPY --from=go-builder /out/nunchucks /usr/local/bin/nunchucks

EXPOSE 3000
CMD ["node", "./app/server.mjs"]
