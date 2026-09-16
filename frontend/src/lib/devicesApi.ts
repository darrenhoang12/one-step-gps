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
      icon_url: z.url().nullable().optional(),
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
      icon_url: device.icon_url ?? null,
    };
  });
}

export async function updateDevicePreferences(
  update: DevicePreferenceUpdate,
): Promise<void> {
  const response = await fetch(`${env.apiBaseUrl}/preferences`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(update),
  });
  if (!response.ok)
    throw new Error(`Could not save device preferences (${response.status})`);
}

export async function updateDeviceOrder(deviceIds: string[]): Promise<void> {
  const response = await fetch(`${env.apiBaseUrl}/preferences/order`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ device_ids: deviceIds }),
  });
  if (!response.ok)
    throw new Error(`Could not save device order (${response.status})`);
}

export async function uploadDeviceIcon(
  deviceId: string,
  file: File,
): Promise<{ iconStoragePath: string; iconUrl: string }> {
  const maxIconBytes = 5 * 1024 * 1024;
  const maxIconDimension = 1024;
  if (file.size > maxIconBytes) {
    throw new Error("Icon must be 5 MB or smaller");
  }

  const dimensions = await readImageDimensions(file);
  if (
    dimensions.width > maxIconDimension ||
    dimensions.height > maxIconDimension
  ) {
    throw new Error("Icon dimensions must be 1024×1024 pixels or smaller");
  }

  const form = new FormData();
  form.set("device_id", deviceId);
  form.set("icon", file);
  const response = await fetch(`${env.apiBaseUrl}/preferences/icon`, {
    method: "POST",
    body: form,
  });
  if (!response.ok)
    throw new Error(`Could not upload device icon (${response.status})`);
  const payload = z
    .object({
      icon_storage_path: z.string(),
      icon_url: z.string().url(),
    })
    .parse(await response.json());
  return {
    iconStoragePath: payload.icon_storage_path,
    iconUrl: payload.icon_url,
  };
}

export async function removeDeviceIcon(deviceId: string): Promise<void> {
  const response = await fetch(`${env.apiBaseUrl}/preferences/icon`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ device_id: deviceId }),
  });
  if (!response.ok) {
    throw new Error(`Could not remove device icon (${response.status})`);
  }
}

function readImageDimensions(
  file: File,
): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    const objectUrl = URL.createObjectURL(file);
    image.onload = () => {
      URL.revokeObjectURL(objectUrl);
      resolve({ width: image.naturalWidth, height: image.naturalHeight });
    };
    image.onerror = () => {
      URL.revokeObjectURL(objectUrl);
      reject(new Error("The selected icon is not a valid image"));
    };
    image.src = objectUrl;
  });
}
