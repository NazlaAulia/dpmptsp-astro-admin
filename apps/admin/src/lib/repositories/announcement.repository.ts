// Announcement data access for the admin panel.
import { api, type Announcement, type AnnouncementInput } from "@dpmptsp/api-client";
import { currentSessionId } from "../request-context";

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

// Writes carry the signed-in administrator's session: the service key alone
// only proves the caller is one of our own apps, and the API refuses a write
// without a session.
export async function create(body: AnnouncementInput) {
  return api.createAnnouncement(body, currentSessionId());
}

export async function update(id: number, body: AnnouncementInput) {
  return api.updateAnnouncement(id, body, currentSessionId());
}

export async function remove(id: number) {
  return api.deleteAnnouncement(id, currentSessionId());
}
