-- Tabel untuk Pemberitahuan Admin
CREATE TABLE IF NOT EXISTS pemberitahuan (
    id INT PRIMARY KEY AUTO_INCREMENT,
    judul VARCHAR(255) NOT NULL,
    deskripsi TEXT NOT NULL,
    foto VARCHAR(255),
    link_url VARCHAR(500),
    status_aktif ENUM('Y', 'N') DEFAULT 'Y',
    urutan INT DEFAULT 0,
    tipe ENUM('notif', 'modal') DEFAULT 'notif',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Database sudah siap, upload notifikasi dari admin
-- Data sebelumnya sudah dihapus
