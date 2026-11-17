# Pin the base images by digest in production, e.g.
#   docker build --build-arg NODE_IMAGE=node:24-alpine@sha256:... .
ARG NODE_IMAGE=node:24-alpine
FROM ${NODE_IMAGE} AS build
WORKDIR /src
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web ./
# Copies Vditor's runtime assets into public/vditor so no third-party CDN is
# reachable from the admin session at runtime.
RUN npm run build

ARG NGINX_IMAGE=nginx:1.27-alpine
FROM ${NGINX_IMAGE}
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
