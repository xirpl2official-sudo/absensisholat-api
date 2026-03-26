-- Update Demo Admin Password Script
-- Run this to fix the password hash for the demo admin account
-- Password: Demo@12345

UPDATE users_staff 
SET password = '$2a$10$H.9j0cktgWkWL1dOi7Vld.q/4BDaVH43h1K0KORW5q7wpLwE0YLKW'
WHERE username = 'admin_demo';
