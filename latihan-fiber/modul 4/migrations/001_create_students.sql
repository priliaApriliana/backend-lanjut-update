-- Migrasi 001: Membuat tabel students dan indeks
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks unik untuk NIM tanpa membedakan huruf besar dan kecil.
-- Menjamin keunikan data di tingkat basis data (ACID) dan mencegah race condition.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks tambahan pada nama mahasiswa untuk mempercepat klausa pencarian ILIKE dan ORDER BY
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
