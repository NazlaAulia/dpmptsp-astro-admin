// Announcement data access for the admin panel.
import { api, type Announcement, type AnnouncementInput } from "@dpmptsp/api-client";

export async function findAll(): Promise<Announcement[]> {
  const res = await api.announcements({ includeInactive: true });
  if (!res.ok) {
    console.error("announcements failed:", res.status, res.error);
    return [];
  }
  return res.data;
}

export async function findById(id: number): Promise<Announcement | null> {
  const res = await api.announcement(id);
  return res.ok ? res.data : null;
}

export async function create(body: AnnouncementInput) {
  return api.createAnnouncement(body);
}

export async function update(id: number, body: AnnouncementInput) {
  return api.updateAnnouncement(id, body);
}

export async function remove(id: number) {
  return api.deleteAnnouncement(id);
}
