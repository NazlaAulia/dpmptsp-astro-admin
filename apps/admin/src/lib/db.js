// TEMPORARY. This file is deleted once apps/api (Go) owns database access,
// per CLAUDE.md rule 3: the frontend must never touch the database directly.
//
// Deliberately duplicated in apps/web and apps/admin rather than promoted to
// packages/*. Sharing it would bless an architecture violation as shared
// infrastructure and guarantee it outlives the Go API.
//
// Do not add queries here.
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
