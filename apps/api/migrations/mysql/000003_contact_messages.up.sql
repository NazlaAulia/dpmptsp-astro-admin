-- Contact and complaint messages (MySQL).
--
-- Generated from the Postgres migration.
--
-- Contact and complaint messages.
--
-- ticket_code is the only credential for reading a ticket back, so it is random
-- rather than sequential: a counter would let anyone enumerate other people's
-- complaints.

BEGIN;

CREATE TABLE contact_messages (
    id          INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ticket_code VARCHAR(32)  NOT NULL,
    nama        VARCHAR(255) NOT NULL,
    telepon     VARCHAR(50),
    email       VARCHAR(255),
    kategori    VARCHAR(100),
    subjek      VARCHAR(255),
    pesan       TEXT         NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'baru',
    catatan     TEXT,
    ip_address  VARCHAR(45),
    created_at  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT contact_messages_status_chk
        CHECK (status IN ('baru','diproses','selesai','ditolak')),
    UNIQUE KEY contact_messages_ticket_key (ticket_code),
    KEY contact_messages_status_created_idx (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


COMMIT;
