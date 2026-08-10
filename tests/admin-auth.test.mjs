import test from 'node:test';
import assert from 'node:assert/strict';
import { authenticateAdmin } from '../src/lib/adminAuth.js';

test('login berhasil dengan kredensial bawaan admin', async () => {
  const result = await authenticateAdmin({
    username: 'admin',
    password: 'admin123',
    db: null,
  });

  assert.equal(result.success, true);
  assert.equal(result.error, undefined);
});

test('login gagal untuk kredensial yang salah', async () => {
  const result = await authenticateAdmin({
    username: 'admin',
    password: 'salah',
    db: null,
  });

  assert.equal(result.success, false);
  assert.equal(result.error, 'Username atau password salah!');
});
