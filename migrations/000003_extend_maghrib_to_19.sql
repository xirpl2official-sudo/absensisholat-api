-- +migrate Up
-- Extend Maghrib end time to 19:00 for testing

UPDATE jadwal_sholat
SET waktu_selesai = '19:00:00'
WHERE jenis_sholat = 'Maghrib';

-- +migrate Down
-- Revert Maghrib end time back to 18:30

UPDATE jadwal_sholat
SET waktu_selesai = '18:30:00'
WHERE jenis_sholat = 'Maghrib';
