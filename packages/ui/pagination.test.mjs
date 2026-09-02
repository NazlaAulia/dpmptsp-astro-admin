import test from "node:test";
import assert from "node:assert/strict";
import { pageWindow } from "./src/pagination.ts";

// pageWindow used to be inline arithmetic inside ArtikelTable.astro, duplicated
// with small variations across several listing pages. Pinning it here is what
// makes it safe to reuse.

test("shows every page when there are fewer than five", () => {
  const w = pageWindow(1, 3, 25, 10);
  assert.deepEqual(w.visiblePages, [1, 2, 3]);
});

test("caps the window at five pages", () => {
  const w = pageWindow(10, 40, 400, 10);
  assert.equal(w.visiblePages.length, 5);
});

test("centres the window on the current page", () => {
  const w = pageWindow(10, 40, 400, 10);
  assert.deepEqual(w.visiblePages, [8, 9, 10, 11, 12]);
});

test("does not run off the start", () => {
  const w = pageWindow(1, 40, 400, 10);
  assert.deepEqual(w.visiblePages, [1, 2, 3, 4, 5]);
});

test("does not run off the end", () => {
  const w = pageWindow(40, 40, 400, 10);
  assert.deepEqual(w.visiblePages, [36, 37, 38, 39, 40]);
});

test("reports the showing range", () => {
  const w = pageWindow(2, 4, 35, 10);
  assert.equal(w.showingStart, 11);
  assert.equal(w.showingEnd, 20);
});

test("the last page does not claim more rows than exist", () => {
  const w = pageWindow(4, 4, 35, 10);
  assert.equal(w.showingStart, 31);
  assert.equal(w.showingEnd, 35, "showingEnd must clamp to the real total");
});

test("an empty table shows 0 to 0", () => {
  const w = pageWindow(1, 1, 0, 10);
  assert.equal(w.showingStart, 0);
  assert.equal(w.showingEnd, 0);
});
