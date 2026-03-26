-- Insert jadwal_sholat for all prayer times (all days)
-- This ensures QR code generation has active schedules to work with

-- First, clear existing entries (optional, comment out if you want to preserve existing data)
-- DELETE FROM jadwal_sholat;

-- Insert Subuh (early morning)
INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Senin', 'Subuh', '04:30:00', '05:30:00', NULL),
('Selasa', 'Subuh', '04:30:00', '05:30:00', NULL),
('Rabu', 'Subuh', '04:30:00', '05:30:00', NULL),
('Kamis', 'Subuh', '04:30:00', '05:30:00', NULL),
('Jumat', 'Subuh', '04:30:00', '05:30:00', NULL),
('Sabtu', 'Subuh', '04:30:00', '05:30:00', NULL),
('Minggu', 'Subuh', '04:30:00', '05:30:00', NULL);

-- Insert Dzuhur (midday)
INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Senin', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Selasa', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Rabu', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Kamis', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Jumat', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Sabtu', 'Dzuhur', '12:00:00', '13:00:00', NULL),
('Minggu', 'Dzuhur', '12:00:00', '13:00:00', NULL);

-- Insert Asr (afternoon)
INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Senin', 'Asr', '15:00:00', '16:00:00', NULL),
('Selasa', 'Asr', '15:00:00', '16:00:00', NULL),
('Rabu', 'Asr', '15:00:00', '16:00:00', NULL),
('Kamis', 'Asr', '15:00:00', '16:00:00', NULL),
('Jumat', 'Asr', '15:00:00', '16:00:00', NULL),
('Sabtu', 'Asr', '15:00:00', '16:00:00', NULL),
('Minggu', 'Asr', '15:00:00', '16:00:00', NULL);

-- Insert Maghrib (evening) - Current active time
INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Senin', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Selasa', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Rabu', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Kamis', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Jumat', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Sabtu', 'Maghrib', '17:30:00', '18:30:00', NULL),
('Minggu', 'Maghrib', '17:30:00', '18:30:00', NULL);

-- Insert Isya (night)
INSERT INTO jadwal_sholat (hari, jenis_sholat, waktu_mulai, waktu_selesai, jurusan) VALUES
('Senin', 'Isya', '19:00:00', '20:00:00', NULL),
('Selasa', 'Isya', '19:00:00', '20:00:00', NULL),
('Rabu', 'Isya', '19:00:00', '20:00:00', NULL),
('Kamis', 'Isya', '19:00:00', '20:00:00', NULL),
('Jumat', 'Isya', '19:00:00', '20:00:00', NULL),
('Sabtu', 'Isya', '19:00:00', '20:00:00', NULL),
('Minggu', 'Isya', '19:00:00', '20:00:00', NULL);
