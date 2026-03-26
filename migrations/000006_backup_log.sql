-- +migrate Up
-- Create backup_log table to track export/backup status

CREATE TABLE backup_log (
    id SERIAL PRIMARY KEY,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    file_format VARCHAR(10) NOT NULL DEFAULT 'xlsx',
    exported_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    auto_delete_after TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by INTEGER REFERENCES users_staff(id_staff)
);

CREATE INDEX idx_backup_log_dates ON backup_log(start_date, end_date);
CREATE INDEX idx_backup_log_exported ON backup_log(exported_at);

-- +migrate Down
DROP TABLE IF EXISTS backup_log;
