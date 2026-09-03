// Admin content data access, through the Go API.

import { currentSessionId } from "../request-context";
import { api } from "@dpmptsp/api-client";
import type { AboutContent, Innovation, PerformanceDoc, ServiceLocation } from "@dpmptsp/api-client";

export type Result<T> = { ok: true; data: T } | { ok: false; message: string };

function fail(name: string, status: number, error: string): { ok: false; message: string } {
  console.error(`${name} failed:`, status, error);
  return {
    ok: false,
    message: status === 404 ? "Data tidak ditemukan." : "Terjadi kesalahan pada server.",
  };
}

async function one<T>(
  name: string,
  call: () => Promise<{ ok: true; data: T } | { ok: false; status: number; error: string }>
): Promise<Result<T>> {
  const res = await call();
  return res.ok ? { ok: true, data: res.data } : fail(name, res.status, res.error);
}

async function many<T>(
  name: string,
  call: () => Promise<{ ok: true; data: T[] } | { ok: false; status: number; error: string }>
): Promise<T[]> {
  const res = await call();
  if (!res.ok) {
    console.error(`${name} failed:`, res.status, res.error);
    return [];
  }
  return res.data;
}

// --- innovations ---
export const findInnovations = () => many<Innovation>("innovations", () => api.innovations());
export const findInnovation = (id: number) => one<Innovation>("innovation", () => api.innovation(id));
export const createInnovation = (b: Partial<Innovation>) => one<Innovation>("createInnovation", () => api.createInnovation(b, currentSessionId()));
export const updateInnovation = (id: number, b: Partial<Innovation>) => one<Innovation>("updateInnovation", () => api.updateInnovation(id, b, currentSessionId()));
export const deleteInnovation = (id: number) => one<null>("deleteInnovation", () => api.deleteInnovation(id, currentSessionId()));

// --- performance docs (kinerja) ---
export const findPerformanceDocs = () => many<PerformanceDoc>("performanceDocs", () => api.performanceDocs());
export const findPerformanceDoc = (id: number) => one<PerformanceDoc>("performanceDoc", () => api.performanceDoc(id));
export const createPerformanceDoc = (b: Partial<PerformanceDoc>) => one<PerformanceDoc>("createPerformanceDoc", () => api.createPerformanceDoc(b, currentSessionId()));
export const updatePerformanceDoc = (id: number, b: Partial<PerformanceDoc>) => one<PerformanceDoc>("updatePerformanceDoc", () => api.updatePerformanceDoc(id, b, currentSessionId()));
export const deletePerformanceDoc = (id: number) => one<null>("deletePerformanceDoc", () => api.deletePerformanceDoc(id, currentSessionId()));

// --- service locations (layanan) ---
export const findServiceLocations = () => many<ServiceLocation>("serviceLocations", () => api.serviceLocations());
export const findServiceLocation = (id: number) => one<ServiceLocation>("serviceLocation", () => api.serviceLocation(id));
export const createServiceLocation = (b: Partial<ServiceLocation>) => one<ServiceLocation>("createServiceLocation", () => api.createServiceLocation(b, currentSessionId()));
export const updateServiceLocation = (id: number, b: Partial<ServiceLocation>) => one<ServiceLocation>("updateServiceLocation", () => api.updateServiceLocation(id, b, currentSessionId()));
export const deleteServiceLocation = (id: number) => one<null>("deleteServiceLocation", () => api.deleteServiceLocation(id, currentSessionId()));

// --- about contents (konten) ---
export const findAboutContents = () => many<AboutContent>("aboutContents", () => api.aboutContents());
export const findAboutContent = (id: number) => one<AboutContent>("aboutContent", () => api.aboutContent(id));
export const createAboutContent = (b: Partial<AboutContent>) => one<AboutContent>("createAboutContent", () => api.createAboutContent(b, currentSessionId()));
export const updateAboutContent = (id: number, b: Partial<AboutContent>) => one<AboutContent>("updateAboutContent", () => api.updateAboutContent(id, b, currentSessionId()));
export const deleteAboutContent = (id: number) => one<null>("deleteAboutContent", () => api.deleteAboutContent(id, currentSessionId()));
