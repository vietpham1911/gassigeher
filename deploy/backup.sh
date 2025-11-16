#!/bin/bash
# Gassigeher Database Backup Script
# Run daily via cron: 0 2 * * * /var/gassigeher/deploy/backup.sh

set -e

# Configuration
DB_PATH="/var/gassigeher/data/gassigeher.db"
BACKUP_DIR="/var/gassigeher/backups"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/gassigeher_${DATE}.db"
LOG_FILE="/var/gassigeher/logs/backup.log"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Log start
echo "[$(date)] Starting backup..." >> "$LOG_FILE"

# Create backup
if sqlite3 "$DB_PATH" ".backup '$BACKUP_FILE'"; then
    echo "[$(date)] Backup created: $BACKUP_FILE" >> "$LOG_FILE"

    # Compress backup
    gzip "$BACKUP_FILE"
    echo "[$(date)] Backup compressed: ${BACKUP_FILE}.gz" >> "$LOG_FILE"

    # Calculate size
    SIZE=$(du -h "${BACKUP_FILE}.gz" | cut -f1)
    echo "[$(date)] Backup size: $SIZE" >> "$LOG_FILE"
else
    echo "[$(date)] ERROR: Backup failed!" >> "$LOG_FILE"
    exit 1
fi

# Remove old backups (older than RETENTION_DAYS)
find "$BACKUP_DIR" -name "gassigeher_*.db.gz" -type f -mtime +$RETENTION_DAYS -delete
DELETED=$(find "$BACKUP_DIR" -name "gassigeher_*.db.gz" -type f -mtime +$RETENTION_DAYS 2>/dev/null | wc -l)
if [ "$DELETED" -gt 0 ]; then
    echo "[$(date)] Removed $DELETED old backup(s)" >> "$LOG_FILE"
fi

# Count remaining backups
TOTAL=$(find "$BACKUP_DIR" -name "gassigeher_*.db.gz" -type f | wc -l)
echo "[$(date)] Total backups: $TOTAL" >> "$LOG_FILE"
echo "[$(date)] Backup completed successfully" >> "$LOG_FILE"

# Optional: Upload to remote storage (uncomment and configure)
# rsync -az "${BACKUP_FILE}.gz" user@backup-server:/backups/gassigeher/

exit 0
