#!/usr/bin/env bash
set -euo pipefail

# PidieKab Automated Release Deployment Script (Ubuntu 24.04 Target)
# Follows atomic release switching and non-root execution principles.

APP_DIR="/opt/pidiekab"
RELEASE_TAG=$(date +%Y%m%d%H%M%S)
NEW_RELEASE_DIR="${APP_DIR}/releases/${RELEASE_TAG}"

echo "==> Deploying PidieKab release: ${RELEASE_TAG}"

# 1. Prepare release directory
mkdir -p "${NEW_RELEASE_DIR}/public"
mkdir -p "${NEW_RELEASE_DIR}/cms"
mkdir -p "${NEW_RELEASE_DIR}/bin"

# 2. Copy artifacts
echo "==> Copying build artifacts..."
cp -r frontend/public-site/dist/* "${NEW_RELEASE_DIR}/public/"
cp -r frontend/admin-cms/dist/* "${NEW_RELEASE_DIR}/cms/"
cp bin/pidiekab "${NEW_RELEASE_DIR}/bin/"
chmod +x "${NEW_RELEASE_DIR}/bin/pidiekab"

# 3. Apply database migrations
echo "==> Applying database migrations with Goose..."
goose -dir migrations postgres "${PIDIEKAB_DATABASE_URL}" up

# 4. Atomic symlink switch
echo "==> Switching active release symlink..."
ln -sfn "${NEW_RELEASE_DIR}" "${APP_DIR}/current"

# 5. Restart systemd service
echo "==> Restarting pidiekab.service..."
systemctl restart pidiekab.service

echo "==> Deployment completed successfully: ${RELEASE_TAG}"
