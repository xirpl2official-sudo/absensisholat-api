-- +migrate Up
-- Insert default jadwal_sholat entries for all prayer times

INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
-- Subuh (Early morning: 04:30-05:30)
('Senin', 'Subuh', '04:30:00', '05:30:00', NULL),
('Selasa', 'Subuh', '04:30:00', '05:30:00', NULL),
('Rabu', 'Subuh', '04:30:00', '05:30:00', NULL),
('Kamis', 'Subuh', '04:30:00', '05:30:00', NULL),
('Jumat', 'Subuh', '04:30:00', '05:30:00', NULL),
('Sabtu', 'Subuh', '04:30:00', '05:30:00', NULL),
('Minggu', 'Subuh', '04:30:00', '05:30:00', NULL),

-- Dzuhur (Midday: 12:00-13:00)
('Senin', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Selasa', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Rabu', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Kamis', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Jumat', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Sabtu', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Minggu', 'Dzuhur', '12:00:00', '13:00:00', NULL),

-- Asr (Afternoon: 15:00-16:00)
('Senin', 'Asr', '15:00:00', '16:00:00', NULL),
('Selasa', 'Asr', '15:00:00', '16:00:00', NULL),
('Rabu', 'Asr', '15:00:00', '16:00:00', NULL),
('Kamis', 'Asr', '15:00:00', '16:00:00', NULL),
('Jumat', 'Asr', '15:00:00', '16:00:00', NULL),
('Sabtu', 'Asr', '15:00:00', '16:00:00', NULL),
('Minggu', 'Asr', '15:00:00', '16:00:00', NULL),

-- Maghrib (Evening: 17:30-18:30)
('Senin', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Selasa', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Rabu', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Kamis', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Jumat', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Sabtu', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Minggu', 'Maghrib', '17:30:00', '18:30:00', NULL),

-- Isya (Night: 19:00-20:00)
('Senin', 'Isya', '19:00:00', '20:00:00', NULL),
('Selasa', 'Isya', '19:00:00', '20:00:00', NULL),
('Rabu', 'Isya', '19:00:00', '20:00:00', NULL),
('Kamis', 'Isya', '19:00:00', '20:00:00', NULL),
('Jumat', 'Isya', '19:00:00', '20:00:00', NULL),
('Sabtu', 'Isya', '19:00:00', '20:00:00', NULL),
('Minggu', 'Isya', '19:00:00', '20:00:00', NULL);

-- +migrate Down
DELETE FROM jadwal_sholat WHERE jenis_sholat IN ('Subuh', 'Dzuhur', 'Asr', 'Maghrib', 'Isya');
