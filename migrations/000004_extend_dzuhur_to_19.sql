-- +migrate Up
-- Extend Dzuhur end time to 19:00 for testing

UPDATE jadwal_sholat
SET waktu_selesai = '19:00:00'
WHERE jenis_sholat = 'Dzuhur';

-- +migrate Down
-- Revert Dzuhur end time back to 13:00

UPDATE jadwal_sholat
SET waktu_selesai = '13:00:00'
WHERE jenis_sholat = 'Dzuhur';
