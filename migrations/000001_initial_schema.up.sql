-- Migration: Initial Schema
-- Version: 000001
-- Description: Create initial database tables for absensholat-api

-- Users Staff table (admin, guru, wali_kelas)
CREATE TABLE IF NOT EXISTS users_staff (
    id_staff SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Siswa (Students) table
CREATE TABLE IF NOT EXISTS siswa (
    nis VARCHAR(20) PRIMARY KEY,
    nama_siswa VARCHAR(255) NOT NULL,
    jk VARCHAR(10) NOT NULL,
    jurusan VARCHAR(100),
    kelas VARCHAR(50)
);

-- Admin table
CREATE TABLE IF NOT EXISTS admin (
    id_admin SERIAL PRIMARY KEY,
    id_staff INTEGER NOT NULL UNIQUE,
    nama_admin VARCHAR(255) NOT NULL,
    FOREIGN KEY (id_staff) REFERENCES users_staff(id_staff) ON DELETE CASCADE
);

-- Guru (Teacher) table
CREATE TABLE IF NOT EXISTS guru (
    id_guru SERIAL PRIMARY KEY,
    id_staff INTEGER NOT NULL UNIQUE,
    nip VARCHAR(20) NOT NULL UNIQUE,
    nama_guru VARCHAR(255) NOT NULL,
    kelas_wali VARCHAR(50),
    FOREIGN KEY (id_staff) REFERENCES users_staff(id_staff) ON DELETE CASCADE
);

-- Jadwal Sholat (Prayer Schedule) table
CREATE TABLE IF NOT EXISTS jadwal_sholat (
    id_jadwal SERIAL PRIMARY KEY,
    hari VARCHAR(20) NOT NULL,
    jenis_sholat VARCHAR(50) NOT NULL,
    waktu_mulai TIME NOT NULL,
    waktu_selesai TIME NOT NULL,
    jurusan VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sesi Sholat (Prayer Session) table
CREATE TABLE IF NOT EXISTS sesi_sholat (
    id_sholat SERIAL PRIMARY KEY,
    nama_sholat VARCHAR(255) NOT NULL,
    waktu_mulai TIME NOT NULL,
    waktu_selesai TIME NOT NULL,
    tanggal DATE NOT NULL
);

-- Akun Login Siswa (Student Login Account) table
CREATE TABLE IF NOT EXISTS akun_login_siswa (
    nis VARCHAR(20) PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nis) REFERENCES siswa(nis) ON DELETE CASCADE
);

-- Absensi (Attendance) table
CREATE TABLE IF NOT EXISTS absensi (
    id_absen SERIAL PRIMARY KEY,
    nis VARCHAR(20) NOT NULL,
    id_jadwal INTEGER NOT NULL,
    tanggal DATE NOT NULL,
    status VARCHAR(50) NOT NULL,
    deskripsi TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nis) REFERENCES siswa(nis) ON DELETE CASCADE,
    FOREIGN KEY (id_jadwal) REFERENCES jadwal_sholat(id_jadwal) ON DELETE CASCADE
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_absensi_nis ON absensi(nis);
CREATE INDEX IF NOT EXISTS idx_absensi_id_jadwal ON absensi(id_jadwal);
CREATE INDEX IF NOT EXISTS idx_absensi_tanggal ON absensi(tanggal);
CREATE INDEX IF NOT EXISTS idx_jadwal_sholat_hari ON jadwal_sholat(hari);
CREATE INDEX IF NOT EXISTS idx_jadwal_sholat_jenis ON jadwal_sholat(jenis_sholat);
CREATE INDEX IF NOT EXISTS idx_sesi_sholat_tanggal ON sesi_sholat(tanggal);
CREATE INDEX IF NOT EXISTS idx_guru_nip ON guru(nip);
CREATE INDEX IF NOT EXISTS idx_users_staff_username ON users_staff(username);
