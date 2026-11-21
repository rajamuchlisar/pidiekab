#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# PidieKab Automated Release Deployment Script (Ubuntu 24.04 Target)
# Architecture: Single Go Binary, Vue 3 SSG & SPA, PostgreSQL, systemd
# ==============================================================================

APP_DIR="${PIDIEKAB_APP_DIR:-/opt/pidiekab}"
RELEASE_TAG=$(date +%Y%m%d_%H%M%S)
NEW_RELEASE_DIR="${APP_DIR}/releases/${RELEASE_TAG}"
CURRENT_LINK="${APP_DIR}/current"
SERVICE_NAME="pidiekab.service"
HEALTH_URL="${PIDIEKAB_HEALTH_URL:-http://127.0.0.1:8080/health/ready}"

echo "=================================================="
echo "Starting PidieKab Deployment: ${RELEASE_TAG}"
echo "Target Release Directory: ${NEW_RELEASE_DIR}"
echo "=================================================="

# 1. Validate environment & dependencies
if [ -z "${PIDIEKAB_DATABASE_URL:-}" ] && [ -f "/etc/pidiekab/pidiekab.env" ]; then
    PIDIEKAB_DATABASE_URL=$(grep -E '^[[:space:]]*PIDIEKAB_DATABASE_URL=' /etc/pidiekab/pidiekab.env | cut -d '=' -f2- | tr -d '"' | tr -d "'")
fi

if [ -z "${PIDIEKAB_DATABASE_URL:-}" ]; then
    echo "ERROR: PIDIEKAB_DATABASE_URL is not set and could not be loaded from /etc/pidiekab/pidiekab.env." >&2
    exit 1
fi

if [ ! -f "bin/pidiekab" ]; then
    echo "ERROR: Compiled binary bin/pidiekab not found. Build artifact first." >&2
    exit 1
fi

# 2. Record previous release for rollback capability
PREV_RELEASE=""
if [ -L "${CURRENT_LINK}" ]; then
    PREV_RELEASE=$(readlink -f "${CURRENT_LINK}")
    echo "--> Current active release identified: ${PREV_RELEASE}"
fi

# 3. Prepare release directory layout
echo "--> Creating release directory structure..."
mkdir -p "${NEW_RELEASE_DIR}/bin"
mkdir -p "${NEW_RELEASE_DIR}/public"
mkdir -p "${NEW_RELEASE_DIR}/cms"
mkdir -p "${NEW_RELEASE_DIR}/deploy"

# 4. Copy build artifacts
echo "--> Deploying binary and frontend assets..."
cp bin/pidiekab "${NEW_RELEASE_DIR}/bin/pidiekab"
chmod +x "${NEW_RELEASE_DIR}/bin/pidiekab"

if [ -d "frontend/admin-cms/dist" ]; then
    cp -r frontend/admin-cms/dist/* "${NEW_RELEASE_DIR}/cms/"
fi

if [ -d "deploy" ]; then
    cp -r deploy/* "${NEW_RELEASE_DIR}/deploy/"
fi

if [ -d "api" ]; then
    mkdir -p "${NEW_RELEASE_DIR}/api"
    cp -r api/* "${NEW_RELEASE_DIR}/api/"
fi

# Set proper ownership
chown -R pidiekab:pidiekab "${NEW_RELEASE_DIR}"

# 5. Apply database migrations
if command -v goose >/dev/null 2>&1; then
    echo "--> Applying database migrations with Goose..."
    goose -dir migrations postgres "${PIDIEKAB_DATABASE_URL}" up
else
    echo "WARNING: goose command not found in PATH; skipping automatic migration (ensure schema is up-to-date)."
fi

# 6. Atomic symlink switch
echo "--> Switching active release symlink to ${NEW_RELEASE_DIR}..."
ln -sfn "${NEW_RELEASE_DIR}" "${CURRENT_LINK}"

# 7. Restart systemd service
echo "--> Restarting ${SERVICE_NAME}..."
systemctl restart "${SERVICE_NAME}"

# 8. Post-deployment Health & Readiness verification
echo "--> Waiting for service readiness (${HEALTH_URL})..."
READY=0
for i in {1..15}; do
    if curl -s -f "${HEALTH_URL}" > /dev/null 2>&1; then
        READY=1
        break
    fi
    sleep 2
done

if [ "${READY}" -ne 1 ]; then
    echo "ERROR: Readiness probe failed after restart! Initiating automated rollback..." >&2
    if [ -n "${PREV_RELEASE}" ] && [ -d "${PREV_RELEASE}" ]; then
        ln -sfn "${PREV_RELEASE}" "${CURRENT_LINK}"
        systemctl restart "${SERVICE_NAME}"
        echo "--> Rollback to ${PREV_RELEASE} completed."
    fi
    exit 1
fi

echo "--> Service is healthy and ready."

# 9. Prune old releases (Keep last 5 releases)
echo "--> Pruning old releases in ${APP_DIR}/releases..."
cd "${APP_DIR}/releases"
ls -dt */ | tail -n +6 | xargs -r rm -rf

echo "=================================================="
echo "Deployment of ${RELEASE_TAG} SUCCEEDED at $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "=================================================="
