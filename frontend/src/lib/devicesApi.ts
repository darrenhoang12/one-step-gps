import { z } from "zod";
import { env } from "@/lib/env";
import type { Device, DevicePreferenceUpdate } from "@/types/device";

const deviceResponseSchema = z.object({
  result_list: z.array(
    z.object({
      device_id: z.string(),
      display_name: z.string(),
      model: z.string(),
      online: z.boolean(),
      active_state: z.string(),
      sort_order: z.number().int().nonnegative().nullable().optional(),
      archived: z.boolean().optional(),
      hidden: z.boolean().optional(),
      custom_display_name: z.string().nullable().optional(),
      icon_storage_path: z.string().nullable().optional(),
      latest_device_point: z
        .object({
          lat: z.number().nullable().optional(),
          lng: z.number().nullable().optional(),
          speed: z.number().nullable().optional(),
          dt_tracker: z.string().nullable().optional(),
          dt_server: z.string().nullable().optional(),
        })
        .nullable()
        .optional(),
    }),
  ),
});

export async function fetchDevices(signal?: AbortSignal): Promise<Device[]> {
  const response = await fetch(`${env.apiBaseUrl}/get-devices`, { signal });
  if (!response.ok) {
    throw new Error(`Could not load devices (${response.status})`);
  }

  const parsed = deviceResponseSchema.safeParse(await response.json());
  if (!parsed.success) {
    throw new Error("The server returned an unexpected device response");
  }
  const payload = parsed.data;
  return payload.result_list.map((device) => {
    const point = device.latest_device_point;
    const customName = device.custom_display_name?.trim() || null;
    const iconPath = device.icon_storage_path ?? null;
    return {
      device_id: device.device_id,
      display_name: customName ?? device.display_name,
      original_display_name: device.display_name,
      model: device.model,
      online: device.online,
      active_state: device.active_state,
      lat: point?.lat ?? null,
      lng: point?.lng ?? null,
      // The upstream point speed is in km/h; cards display mph.
      speed: (point?.speed ?? 0) / 1.609344,
      last_reported_at: point?.dt_tracker ?? point?.dt_server ?? null,
      sort_order: device.sort_order ?? null,
      archived: device.archived ?? false,
      hidden: device.hidden ?? false,
      custom_display_name: customName,
      icon_storage_path: iconPath,
      icon_url: iconPath ? `${env.apiBaseUrl}${iconPath}` : null,
    };
  });
}

export async function updateDevicePreferences(update: DevicePreferenceUpdate): Promise<void> {
  const response = await fetch(`${env.apiBaseUrl}/preferences`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(update),
  });
  if (!response.ok) throw new Error(`Could not save device preferences (${response.status})`);
}

export async function updateDeviceOrder(deviceIds: string[]): Promise<void> {
  const response = await fetch(`${env.apiBaseUrl}/preferences/order`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ device_ids: deviceIds }),
  });
  if (!response.ok) throw new Error(`Could not save device order (${response.status})`);
}

export async function uploadDeviceIcon(deviceId: string, file: File): Promise<string> {
  const form = new FormData();
  form.set("device_id", deviceId);
  form.set("icon", file);
  const response = await fetch(`${env.apiBaseUrl}/preferences/icon`, {
    method: "POST",
    body: form,
  });
  if (!response.ok) throw new Error(`Could not upload device icon (${response.status})`);
  const payload = z.object({ icon_storage_path: z.string() }).parse(await response.json());
  return payload.icon_storage_path;
}
