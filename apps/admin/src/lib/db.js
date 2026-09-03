// TEMPORARY. This file is deleted once apps/api (Go) owns database access,
// : the frontend must never touch the database directly.
import mysql from "mysql2/promise";
import { optionalEnv, intEnv } from "@dpmptsp/config";

export const db = mysql.createPool({
  host: optionalEnv("DB_HOST", "127.0.0.1"),
  port: intEnv("DB_PORT", 3306),
  user: optionalEnv("DB_USERNAME", "root"),
  password: optionalEnv("DB_PASSWORD", ""),
  database: optionalEnv("DB_DATABASE", "ladpm"),
  waitForConnections: true,
  connectionLimit: intEnv("DB_POOL_SIZE", 10),
  queueLimit: 0,
});
