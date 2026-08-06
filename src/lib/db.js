/*import mysql from 'mysql2/promise';

export const db = await mysql.createConnection({
  host: 'localhost',
  user: 'root',
  password: '',
  database: 'ladpm'
});*/
// src/lib/db.js
import mysql from 'mysql2/promise';

export const db = mysql.createPool({
  host: 'localhost',
  user: 'root',
  password: '',
  database: 'ladpm',
  waitForConnections: true,
  connectionLimit: 10,
  queueLimit: 0,
});

