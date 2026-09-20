-- ---------------------------------------------------------------
-- Permission untuk entity students
-- ---------------------------------------------------------------
INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh mahasiswa'),
    ('student:read:any',   'Melihat data mahasiswa mana pun'),
    ('student:create',     'Menambah data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete',     'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;
-- role 'user' tidak diberi permission students: hak atas data sendiri
-- mengalir dari kepemilikan (owner_id), bukan dari permission.

-- ---------------------------------------------------------------
-- Kolom owner_id: penanda user yang mendaftarkan data mahasiswa
-- ---------------------------------------------------------------
ALTER TABLE students
    ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id);

-- Baris lama belum punya pemilik. Serahkan ke akun admin dengan id terkecil.
-- Bila tidak ada admin, migration dibatalkan dengan pesan yang jelas.
DO $$
DECLARE
    admin_id INTEGER;
BEGIN
    IF EXISTS (SELECT 1 FROM students WHERE owner_id IS NULL) THEN
        SELECT id INTO admin_id FROM users
        WHERE role = 'admin' ORDER BY id LIMIT 1;

        IF admin_id IS NULL THEN
            RAISE EXCEPTION
                'Tidak ada user berrole admin untuk dijadikan pemilik data lama';
        END IF;

        UPDATE students SET owner_id = admin_id WHERE owner_id IS NULL;
    END IF;
END $$;

-- Setelah semua baris terisi, kolom ini boleh dikunci wajib diisi.
ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);