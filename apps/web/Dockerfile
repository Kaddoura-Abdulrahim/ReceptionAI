FROM node:24-alpine AS deps

WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/web/package.json apps/web/package.json
RUN npm ci

FROM node:24-alpine AS build

WORKDIR /app
ARG NEXT_PUBLIC_API_BASE_URL=https://dentaldesk-api.onrender.com
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm --workspace apps/web run build

FROM node:24-alpine AS runtime

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup -S dentaldesk && adduser -S dentaldesk -G dentaldesk
COPY --from=build /app/apps/web/.next/standalone ./
COPY --from=build /app/apps/web/.next/static ./apps/web/.next/static
COPY apps/web/render-start.sh /app/render-start.sh
RUN chmod +x /app/render-start.sh
USER dentaldesk
EXPOSE 3000
CMD ["/app/render-start.sh"]
