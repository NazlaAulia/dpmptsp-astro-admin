# syntax=docker/dockerfile:1.7
#
# One Dockerfile for both Astro apps. Select with --build-arg APP=web|admin.
ARG NODE_VERSION=22.17.1

FROM node:${NODE_VERSION}-alpine AS base
ENV PNPM_HOME=/pnpm PATH=/pnpm:$PATH
RUN corepack enable
WORKDIR /app

# Manifests only, so editing a .astro file does not invalidate the install layer.
FROM base AS deps
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json        apps/web/
COPY apps/admin/package.json      apps/admin/
COPY packages/config/package.json packages/config/
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile

FROM deps AS build
ARG APP
COPY . .
RUN pnpm --filter "./apps/${APP}" build

FROM base AS runtime
ARG APP
ENV NODE_ENV=production HOST=0.0.0.0 PORT=3000
# The whole workspace tree is copied rather than just the app's own
# node_modules: pnpm's isolated layout symlinks into a store at the workspace
# root, so copying the app directory alone leaves dangling links.
COPY --from=build /app /app
WORKDIR /app/apps/${APP}
EXPOSE 3000
USER node
CMD ["node", "./dist/server/entry.mjs"]
