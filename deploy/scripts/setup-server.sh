#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# PidieKab Production Host Provisioning Script (Ubuntu 24.04 LTS Target)
# Sets up system users, security sandboxing directories, systemd unit, and cron
# ==============================================================================

if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: This provisioning script must be run as root (or with sudo)." >&2
    exit 1
fi

echo "=================================================="
echo "Starting PidieKab Production Host Setup (Ubuntu 24.04)"
echo "=================================================="

# 1. Install required system packages
echo "--> [1/7] Installing base system dependencies..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq \
    curl \
    ca-certificates \
    tar \
    gzip \
    coreutils \
    libcap2-bin \
    postgresql-client

# 2. Create system user and group for PidieKab service
echo "--> [2/7] Configuring dedicated system service account..."
if ! id -u pidiekab >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/sbin/nologin --user-group pidiekab
    echo "    Created system user 'pidiekab' (no login shell)"
else
    echo "    System user 'pidiekab' already exists"
fi

# 3. Create directory layout with strict security permissions
echo "--> [3/7] Setting up application filesystem layout..."
mkdir -p /opt/pidiekab/releases
mkdir -p /var/lib/pidiekab/media
mkdir -p /var/lib/pidiekab/ssg
mkdir -p /var/backups/pidiekab
mkdir -p /etc/pidiekab

# Permissions:
# - /opt/pidiekab/releases: pidiekab can write releases
# - /var/lib/pidiekab: persistent media uploads and SSG builds
# - /var/backups/pidiekab: 0700 restricted to backup user
# - /etc/pidiekab: read-only to service, owned by root:pidiekab with 0640/0750
chown -R pidiekab:pidiekab /opt/pidiekab/releases
chmod 0755 /opt/pidiekab
chmod 0750 /opt/pidiekab/releases

chown -R pidiekab:pidiekab /var/lib/pidiekab
chmod -R 0750 /var/lib/pidiekab

chown -R pidiekab:pidiekab /var/backups/pidiekab
chmod 0700 /var/backups/pidiekab

chown -R root:pidiekab /etc/pidiekab
chmod 0750 /etc/pidiekab

echo "    Directory hierarchy created and secured"

# 4. Deploy systemd unit configuration
echo "--> [4/7] Installing systemd service unit..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYSTEMD_SRC="${SCRIPT_DIR}/../systemd/pidiekab.service"

if [ -f "${SYSTEMD_SRC}" ]; then
    cp "${SYSTEMD_SRC}" /etc/systemd/system/pidiekab.service
    chmod 0644 /etc/systemd/system/pidiekab.service
    echo "    Installed /etc/systemd/system/pidiekab.service"
else
    echo "ERROR: Systemd service source not found at ${SYSTEMD_SRC}" >&2
    exit 1
fi

# 5. Deploy environment file template if not existing
echo "--> [5/7] Checking production environment file..."
ENV_SRC="${SCRIPT_DIR}/../systemd/pidiekab.env.example"
ENV_DEST="/etc/pidiekab/pidiekab.env"

if [ ! -f "${ENV_DEST}" ]; then
    if [ -f "${ENV_SRC}" ]; then
        cp "${ENV_SRC}" "${ENV_DEST}"
        chown root:pidiekab "${ENV_DEST}"
        chmod 0640 "${ENV_DEST}"
        echo "    Initialized ${ENV_DEST} from template"
    else
        echo "WARNING: Environment template ${ENV_SRC} not found"
    fi
else
    echo "    Existing ${ENV_DEST} preserved"
fi

# 6. Configure automated daily backup cron job
echo "--> [6/7] Configuring automated daily backup cron..."
BACKUP_SCRIPT_PATH="/opt/pidiekab/current/deploy/scripts/backup.sh"
CRON_FILE="/etc/cron.d/pidiekab-backup"

cat <<'CRON' > "${CRON_FILE}"
# /etc/cron.d/pidiekab-backup
# Daily backup of PostgreSQL database and media storage at 02:00 AM UTC
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

0 2 * * * pidiekab [ -f /opt/pidiekab/current/deploy/scripts/backup.sh ] && [ -f /etc/pidiekab/pidiekab.env ] && set -a && . /etc/pidiekab/pidiekab.env && set +a && /opt/pidiekab/current/deploy/scripts/backup.sh >> /var/log/pidiekab-backup.log 2>&1
CRON

chmod 0644 "${CRON_FILE}"
touch /var/log/pidiekab-backup.log
chown pidiekab:pidiekab /var/log/pidiekab-backup.log
chmod 0644 /var/log/pidiekab-backup.log
echo "    Installed /etc/cron.d/pidiekab-backup (daily at 02:00 UTC)"

# 7. Reload systemd daemon
echo "--> [7/7] Reloading systemd daemon..."
systemctl daemon-reload
systemctl enable pidiekab.service

echo "=================================================="
echo "PidieKab Host Provisioning COMPLETED Successfully!"
echo "=================================================="
echo ""
echo "Next steps for operator:"
echo "1. Edit /etc/pidiekab/pidiekab.env with production database credentials & secrets."
echo "2. Run deployment script: bash deploy/scripts/deploy.sh"
echo "3. Check service status:  systemctl status pidiekab.service"
echo "4. Follow logs:           journalctl -u pidiekab.service -f"
echo ""
