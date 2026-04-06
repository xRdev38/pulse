FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:20-alpine
RUN apk --no-cache add wget
WORKDIR /app
RUN addgroup -S pulse && adduser -S pulse -G pulse
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package.json .
USER pulse
EXPOSE 3000
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/main"]
