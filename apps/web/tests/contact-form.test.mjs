import test from "node:test";
import assert from "node:assert/strict";

const { readForm, validate } = await import("../src/lib/contact-form.ts");

test("reads and trims every field", () => {
  const form = readForm({ nama: "  Budi  ", pesan: " halo ", email: "a@b.co" });
  assert.equal(form.nama, "Budi");
  assert.equal(form.pesan, "halo");
  assert.equal(form.email, "a@b.co");
});

test("coerces non-string fields to empty rather than passing them on", () => {
  // A JSON array or object here would otherwise reach the API as-is.
  const form = readForm({ nama: ["Budi"], pesan: { a: 1 }, telepon: 81234567890 });
  assert.equal(form.nama, "");
  assert.equal(form.pesan, "");
  assert.equal(form.telepon, "");
});

test("survives a null or non-object body", () => {
  for (const body of [null, undefined, "nope", 42]) {
    assert.equal(readForm(body).nama, "");
  }
});

test("caps each field at its own limit", () => {
  const form = readForm({ nama: "a".repeat(500), pesan: "b".repeat(9000) });
  assert.equal(form.nama.length, 120);
  assert.equal(form.pesan.length, 5000);
});

test("requires nama and pesan", () => {
  assert.match(validate(readForm({ pesan: "halo" })), /Nama/);
  assert.match(validate(readForm({ nama: "Budi" })), /pengaduan/i);
  assert.equal(validate(readForm({ nama: "Budi", pesan: "halo" })), null);
});

test("rejects a malformed email but allows an empty one", () => {
  assert.match(validate(readForm({ nama: "B", pesan: "p", email: "bukan-email" })), /email/i);
  assert.equal(validate(readForm({ nama: "B", pesan: "p", email: "" })), null);
});

test("keeps the honeypot value so the API can act on it", () => {
  assert.equal(readForm({ nama: "B", pesan: "p", website: "spam" }).website, "spam");
});
