// Shape and validation for the public contact form.
//
// Kept apart from the service so it depends on nothing: the rules are pure
// functions over an unknown JSON body, which is also what makes them testable
// without an API client or an environment.

/** Longest value accepted for each field. Anything longer is a mistake or a bot. */
const LIMITS = {
  nama: 120,
  telepon: 40,
  email: 160,
  kategori: 80,
  subjek: 200,
  pesan: 5000,
  website: 200,
} as const;

export type ContactForm = { [K in keyof typeof LIMITS]: string };

/**
 * Reads the form out of an unknown JSON body. Every field is coerced to a
 * trimmed, length-capped string, so the API never sees an array or an object
 * where it expects text.
 */
export function readForm(body: unknown): ContactForm {
  const src = (body ?? {}) as Record<string, unknown>;
  const field = (name: keyof typeof LIMITS) => {
    const v = src[name];
    return typeof v === "string" ? v.trim().slice(0, LIMITS[name]) : "";
  };
  return {
    nama: field("nama"),
    telepon: field("telepon"),
    email: field("email"),
    kategori: field("kategori"),
    subjek: field("subjek"),
    pesan: field("pesan"),
    website: field("website"),
  };
}

/** Returns a message to show the sender, or null when the form is usable. */
export function validate(form: ContactForm): string | null {
  if (!form.nama) return "Nama wajib diisi.";
  if (!form.pesan) return "Isi pengaduan wajib diisi.";
  if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    return "Format email tidak valid.";
  }
  return null;
}
