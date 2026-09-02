// Environment access for every workspace package.
//
// Deliberately dependency-free: SPEC.md §11.1 and rule 8 say no new frameworks
// or libraries without a concrete need, and this is thirty lines. It exists so
// configuration fails loudly at boot instead of surfacing as a confusing error
// on the first request.

/** @param {string} name */
function raw(name) {
  const v = process.env[name];
  return v === undefined || v === "" ? undefined : v;
}

/**
 * Required string. Throws at import time if missing.
 * @param {string} name
 * @returns {string}
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
 * @param {string} fallback
 * @returns {string}
 */
export function optionalEnv(name, fallback) {
  return raw(name) ?? fallback;
}

/**
 * Integer with a fallback. Throws if set but not an integer.
 * @param {string} name
 * @param {number} fallback
 * @returns {number}
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
