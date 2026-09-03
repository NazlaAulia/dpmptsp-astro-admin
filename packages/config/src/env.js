// Environment access for every workspace package.

/** @param {string} name */
function raw(name) {
  const v = process.env[name];
  return v === undefined || v === "" ? undefined : v;
}

/**
 * Required string. Throws at import time if missing.
 * @param {string} name
 */
export function requireEnv(name) {
  const v = raw(name);
  if (v === undefined) {
    throw new Error(
      `Missing required environment variable ${name}. ` +
        `Copy .env.example to .env and fill it in.`
    );
  }
  return v;
}

/**
 * Optional string with a fallback.
 * @param {string} name
 */
export function optionalEnv(name, fallback) {
  return raw(name) ?? fallback;
}

/**
 * Integer with a fallback. Throws if set but not an integer.
 * @param {string} name
 */
export function intEnv(name, fallback) {
  const v = raw(name);
  if (v === undefined) return fallback;
  const n = Number(v);
  if (!Number.isInteger(n)) {
    throw new Error(`Environment variable ${name} must be an integer, got "${v}".`);
  }
  return n;
}
