-- Demo Admin Account Insertion Script
-- This script inserts a demo admin account for testing/development
-- 
-- Demo Admin Credentials:
-- Username: admin_demo
-- Password: Demo@12345
-- Hashed Password: $2a$10$H.9j0cktgWkWL1dOi7Vld.q/4BDaVH43h1K0KORW5q7wpLwE0YLKW
-- 
-- NOTE: The hashed password above is a bcrypt hash (cost 10) of "Demo@12345"
-- Change this password in production!

BEGIN;

-- Insert into users_staff table (create login account)
INSERT INTO users_staff (username, password, role, created_at)
VALUES (
    'admin_demo',
    '$2a$10$H.9j0cktgWkWL1dOi7Vld.q/4BDaVH43h1K0KORW5q7wpLwE0YLKW',
    'admin',
    CURRENT_TIMESTAMP
) ON CONFLICT (username) DO NOTHING;

-- Insert into admin table (create admin profile)
-- Get the id_staff from the inserted record
INSERT INTO admin (id_staff, nama_admin)
SELECT id_staff, 'Demo Administrator'
FROM users_staff
WHERE username = 'admin_demo'
AND id_staff NOT IN (SELECT id_staff FROM admin WHERE id_staff IS NOT NULL)
ON CONFLICT (id_staff) DO NOTHING;

COMMIT;

-- Verification Query (run this after insertion to verify):
-- SELECT u.id_staff, u.username, u.role, u.created_at, a.nama_admin
-- FROM users_staff u
-- LEFT JOIN admin a ON u.id_staff = a.id_staff
-- WHERE u.username = 'admin_demo';
