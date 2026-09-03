import test from "node:test";
import assert from "node:assert/strict";
import { slugify } from "../src/utils/slugify.js";

// These tests pin the CURRENT behaviour of slugify, quirks included.

test("lowercases and joins words with hyphens", () => {
  assert.equal(slugify("Kunjungan Wisatawan Kapal Pesiar"), "kunjungan-wisatawan-kapal-pesiar");
});

test("collapses runs of whitespace and trims", () => {
  assert.equal(slugify("  Spasi   Ganda  "), "spasi-ganda");
});

test("drops punctuation", () => {
  assert.equal(slugify("Investasi, Tahun 2024!"), "investasi-tahun-2024");
});

test("keeps underscores, because \\w includes them", () => {
  assert.equal(slugify("Tanda_Bawah"), "tanda_bawah");
});

test("returns empty string for empty and nullish input", () => {
  assert.equal(slugify(""), "");
  assert.equal(slugify(null), "");
  assert.equal(slugify(undefined), "");
});

// --- quirks that will bite the Go port ---

test("QUIRK: an en dash becomes a double hyphen, not a single one", () => {
  // The dash is surrounded by spaces, so both spaces become hyphens and the
  // dash itself is then stripped, leaving two. The legacy article dump
  // contains 20 en dashes, so this shape exists in real URLs.
  assert.equal(slugify("Perizinan Berusaha – OSS"), "perizinan-berusaha--oss");
});

test("QUIRK: accented letters are deleted rather than transliterated", () => {
  // "É" is not in \w, so it is removed outright — the letter is lost, not
  // folded to "e". Any reimplementation that transliterates would produce a
  // different slug and break the existing link.
  assert.equal(slugify("Éksport Impor"), "ksport-impor");
});

test("QUIRK: an ampersand also collapses to a double hyphen", () => {
  assert.equal(slugify("DPMPTSP & Pelayanan"), "dpmptsp--pelayanan");
});
