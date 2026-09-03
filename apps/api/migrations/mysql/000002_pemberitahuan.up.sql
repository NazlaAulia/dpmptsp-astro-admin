-- Announcements (MySQL). Generated from the Postgres migration.
--
-- deskripsi is nullable: the admin screens never supply it.
-- Overlaps notifikasi_beranda, which drives homepage announcements; the two are
-- kept separate because both hold live data.

BEGIN;

CREATE TABLE pemberitahuan (
    id         INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    judul      VARCHAR(255) NOT NULL,
    deskripsi  TEXT,
    foto       VARCHAR(255),
    link_url   VARCHAR(500),
    is_active  TINYINT(1)     NOT NULL DEFAULT 1,
    urutan     INTEGER     NOT NULL DEFAULT 0,
    tipe       VARCHAR(10) NOT NULL DEFAULT 'notif',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pemberitahuan_tipe_chk CHECK (tipe IN ('notif','modal'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Both read paths filter on tipe and is_active, then order.
CREATE INDEX pemberitahuan_tipe_active_idx ON pemberitahuan (tipe, is_active, urutan);

COMMIT;
