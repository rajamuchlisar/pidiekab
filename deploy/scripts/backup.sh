#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# PidieKab Automated Database & Media Backup Script
# Target: Ubuntu 24.04 LTS (PostgreSQL 17 + Local Media Storage)
# ==============================================================================

BACKUP_DIR="${BACKUP_DIR:-/var/backups/pidiekab}"
DATE_TAG=$(date +%Y%m%d_%H%M%S)
DB_BACKUP_FILE="${BACKUP_DIR}/pidiekab_db_${DATE_TAG}.dump"
MEDIA_BACKUP_FILE="${BACKUP_DIR}/pidiekab_media_${DATE_TAG}.tar.gz"
STORAGE_PATH="${PIDIEKAB_STORAGE_PATH:-/var/lib/pidiekab/media}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Ensure backup directory exists with restricted permissions
mkdir -p "${BACKUP_DIR}"
chmod 0700 "${BACKUP_DIR}"

echo "=================================================="
echo "Starting PidieKab Backup: ${DATE_TAG}"
echo "Backup Destination: ${BACKUP_DIR}"
echo "=================================================="

# 1. PostgreSQL Logical Dump
if [ -z "${PIDIEKAB_DATABASE_URL:-}" ] && [ -f "/etc/pidiekab/pidiekab.env" ]; then
    PIDIEKAB_DATABASE_URL=$(grep -E '^[[:space:]]*PIDIEKAB_DATABASE_URL=' /etc/pidiekab/pidiekab.env | cut -d '=' -f2- | tr -d '"' | tr -d "'")
fi

if [ -z "${PIDIEKAB_DATABASE_URL:-}" ]; then
    echo "ERROR: PIDIEKAB_DATABASE_URL is not set and could not be loaded from /etc/pidiekab/pidiekab.env." >&2
    exit 1
fi

echo "--> Dumping PostgreSQL database (custom format)..."
pg_dump --format=custom --no-owner --no-acl \
    --file="${DB_BACKUP_FILE}" \
    "${PIDIEKAB_DATABASE_URL}"

# Verify database dump file exists and is not empty
if [ ! -s "${DB_BACKUP_FILE}" ]; then
    echo "ERROR: Database dump file is empty or was not created." >&2
    exit 1
fi

DB_SIZE=$(du -h "${DB_BACKUP_FILE}" | cut -f1)
echo "--> Database dump successful. Size: ${DB_SIZE}"

# Compute SHA-256 checksum for database backup integrity
sha256sum "${DB_BACKUP_FILE}" > "${DB_BACKUP_FILE}.sha256"
echo "--> Checksum generated: $(cat "${DB_BACKUP_FILE}.sha256")"

# 2. Media Storage Archive
if [ -d "${STORAGE_PATH}" ]; then
    echo "--> Archiving media storage from ${STORAGE_PATH}..."
    tar -czf "${MEDIA_BACKUP_FILE}" -C "$(dirname "${STORAGE_PATH}")" "$(basename "${STORAGE_PATH}")"
    
    if [ ! -s "${MEDIA_BACKUP_FILE}" ]; then
        echo "ERROR: Media archive is empty or failed to create." >&2
        exit 1
    fi
    MEDIA_SIZE=$(du -h "${MEDIA_BACKUP_FILE}" | cut -f1)
    echo "--> Media archive successful. Size: ${MEDIA_SIZE}"
    sha256sum "${MEDIA_BACKUP_FILE}" > "${MEDIA_BACKUP_FILE}.sha256"
else
    echo "WARNING: Media storage path ${STORAGE_PATH} not found; skipping media archive."
fi

# 3. Retention Cleanup (Purge backups older than RETENTION_DAYS)
echo "--> Pruning backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "pidiekab_db_*.dump*" -mtime "+${RETENTION_DAYS}" -delete || true
find "${BACKUP_DIR}" -name "pidiekab_media_*.tar.gz*" -mtime "+${RETENTION_DAYS}" -delete || true

echo "=================================================="
echo "PidieKab Backup completed successfully at $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "=================================================="

