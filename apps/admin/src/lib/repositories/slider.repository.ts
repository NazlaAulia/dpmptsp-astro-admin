// Homepage carousel data access — through the Go API.

import { api, type Slider, type SliderInput } from "@dpmptsp/api-client";
import { currentSessionId } from "../request-context";

export async function findAll(): Promise<Slider[]> {
  const res = await api.sliders();
  if (!res.ok) {
    console.error("sliders failed:", res.status, res.error);
    return [];
  }
  return res.data;
}

export async function findById(id: number): Promise<Slider | null> {
  const res = await api.slider(id);
  return res.ok ? res.data : null;
}

// Writes carry the signed-in administrator's session; the API refuses a write
// without one.
export const create = (body: SliderInput) => api.createSlider(body, currentSessionId());
export const update = (id: number, body: SliderInput) => api.updateSlider(id, body, currentSessionId());
export const remove = (id: number) => api.deleteSlider(id, currentSessionId());
