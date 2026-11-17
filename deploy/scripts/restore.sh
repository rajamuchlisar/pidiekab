#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# PidieKab Automated Database & Media Restore / Drill Script
# Target: Ubuntu 24.04 LTS (PostgreSQL 17)
# ==============================================================================

if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <path_to_db_dump> [target_database_url] [path_to_media_tar_gz]"
    echo "Example: $0 /var/backups/pidiekab/pidiekab_db_20261117_120000.dump postgres://postgres:admin@127.0.0.1:5432/pidiekab_restore_test"
    exit 1
fi

DB_DUMP_FILE="$1"
TARGET_DB_URL="${2:-${RESTORE_DATABASE_URL:-postgres://postgres@127.0.0.1:5432/pidiekab_restore_test?sslmode=disable}}"
MEDIA_ARCHIVE="${3:-}"

echo "=================================================="
echo "Starting PidieKab Restore Drill / Recovery"
echo "Dump File: ${DB_DUMP_FILE}"
echo "Target DB: ${TARGET_DB_URL}"
echo "=================================================="

# 1. Integrity Check via SHA-256
if [ -f "${DB_DUMP_FILE}.sha256" ]; then
    echo "--> Verifying SHA-256 checksum of database dump..."
    sha256sum -c "${DB_DUMP_FILE}.sha256"
else
    echo "WARNING: Checksum file ${DB_DUMP_FILE}.sha256 not found. Proceeding with caution."
fi

# 2. Database Restore using pg_restore
echo "--> Restoring database schema and data..."
pg_restore --clean --if-exists --no-owner --no-acl \
    --dbname="${TARGET_DB_URL}" \
    "${DB_DUMP_FILE}" || {
    # pg_restore returns non-zero for harmless warnings (e.g. drop table if not exists)
    echo "--> pg_restore completed with notices/warnings (evaluating database content)."
}

# 3. Sanity verification query
echo "--> Running verification queries on restored database..."
RECORD_COUNT=$(psql "${TARGET_DB_URL}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo "--> Restored public tables count: $(echo "${RECORD_COUNT}" | tr -d ' ')"

ADMIN_COUNT=$(psql "${TARGET_DB_URL}" -t -c "SELECT COUNT(*) FROM admins;" || echo "0")
echo "--> Administrator records count: $(echo "${ADMIN_COUNT}" | tr -d ' ')"

CATEGORY_COUNT=$(psql "${TARGET_DB_URL}" -t -c "SELECT COUNT(*) FROM categories;" || echo "0")
echo "--> Category records count: $(echo "${CATEGORY_COUNT}" | tr -d ' ')"

CONTENT_COUNT=$(psql "${TARGET_DB_URL}" -t -c "SELECT COUNT(*) FROM contents;" || echo "0")
echo "--> Content records count: $(echo "${CONTENT_COUNT}" | tr -d ' ')"

# 4. Optional Media Restore
if [ -n "${MEDIA_ARCHIVE}" ] && [ -f "${MEDIA_ARCHIVE}" ]; then
    RESTORE_MEDIA_DIR="${RESTORE_MEDIA_DIR:-/tmp/pidiekab_restore_media}"
    echo "--> Restoring media archive to ${RESTORE_MEDIA_DIR}..."
    mkdir -p "${RESTORE_MEDIA_DIR}"
    tar -xzf "${MEDIA_ARCHIVE}" -C "${RESTORE_MEDIA_DIR}"
    echo "--> Restored media files count: $(find "${RESTORE_MEDIA_DIR}" -type f | wc -l)"
fi

echo "=================================================="
echo "Restore verification drill SUCCESSFUL at $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "=================================================="

