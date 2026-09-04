// Homepage carousel rules.

import { create, findAll, findById, remove, update } from "../repositories/slider.repository";
import type { Slider, SliderInput } from "@dpmptsp/api-client";

export type SaveResult = { ok: true; slider: Slider } | { ok: false; message: string };

export const getSliders = findAll;
export const getSlider = findById;

function message(status: number, error: string): string {
  if (status === 422) return error;
  if (status === 401 || status === 403) return "Sesi berakhir. Silakan masuk lagi.";
  return "Gagal menyimpan slide.";
}

export async function createSlider(body: SliderInput): Promise<SaveResult> {
  const res = await create(body);
  return res.ok ? { ok: true, slider: res.data } : { ok: false, message: message(res.status, res.error) };
}

export async function updateSlider(id: number, body: SliderInput): Promise<SaveResult> {
  const res = await update(id, body);
  return res.ok ? { ok: true, slider: res.data } : { ok: false, message: message(res.status, res.error) };
}

export async function deleteSlider(id: number): Promise<{ ok: boolean; message?: string }> {
  const res = await remove(id);
  return res.ok ? { ok: true } : { ok: false, message: message(res.status, res.error) };
}
