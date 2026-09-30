-- Urutan kolom index harus sama persis dengan ORDER BY query,
-- termasuk arah DESC, supaya PostgreSQL bisa memakai index untuk
-- melompat langsung ke posisi cursor.
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);