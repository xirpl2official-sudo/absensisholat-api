-- +migrate Up
-- Add Sholat Jumat entries for Friday
-- Males attend Jumat (11:30-12:30), Females attend Dzuhur (12:30-13:30) on Fridays

INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Jumat', 'Jumat', '11:30:00', '12:30:00', NULL);

-- Update Friday Dzuhur to start after Jumat prayer
UPDATE jadwal_sholat 
SET waktu_mulai = '12:30:00', waktu_selesai = '13:30:00'
WHERE hari = 'Jumat' AND jenis_sholat = 'Dzuhur';

-- +migrate Down
DELETE FROM jadwal_sholat WHERE jenis_sholat = 'Jumat' AND hari = 'Jumat';

UPDATE jadwal_sholat 
SET waktu_mulai = '12:00:00', waktu_selesai = '13:00:00'
WHERE hari = 'Jumat' AND jenis_sholat = 'Dzuhur';
