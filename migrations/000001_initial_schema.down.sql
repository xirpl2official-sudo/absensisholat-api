-- Migration: Initial Schema (Rollback)
-- Version: 000001
-- Description: Remove all initial database tables

DROP INDEX IF EXISTS idx_users_staff_username;
DROP INDEX IF EXISTS idx_guru_nip;
DROP INDEX IF EXISTS idx_sesi_sholat_tanggal;
DROP INDEX IF EXISTS idx_jadwal_sholat_jenis;
DROP INDEX IF EXISTS idx_jadwal_sholat_hari;
DROP INDEX IF EXISTS idx_absensi_tanggal;
DROP INDEX IF EXISTS idx_absensi_id_jadwal;
DROP INDEX IF EXISTS idx_absensi_nis;

DROP TABLE IF EXISTS absensi;
DROP TABLE IF EXISTS akun_login_siswa;
DROP TABLE IF EXISTS sesi_sholat;
DROP TABLE IF EXISTS jadwal_sholat;
DROP TABLE IF EXISTS guru;
DROP TABLE IF EXISTS admin;
DROP TABLE IF EXISTS siswa;
DROP TABLE IF EXISTS users_staff;
