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
  const response = await apiFetch(
    "/get-devices",
    { signal },
    "Could not connect to the server to load devices. Check that the backend is running and try again.",
  );
  await requireSuccessfulResponse(
    response,
    "The server could not load your devices. Please try again.",
  );

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
  const response = await apiFetch(
    "/preferences",
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(update),
    },
    "Could not connect to the server to save this device preference. Check your connection and try again.",
  );
  await requireSuccessfulResponse(
    response,
    "The server could not save this device preference. Please try again.",
  );
}

export async function updateDeviceOrder(deviceIds: string[]): Promise<void> {
  const response = await apiFetch(
    "/preferences/order",
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ device_ids: deviceIds }),
    },
    "Could not connect to the server to save the new device order. Your previous order has been restored.",
  );
  await requireSuccessfulResponse(
    response,
    "The server could not save the new device order. Your previous order has been restored.",
  );
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
  const response = await apiFetch(
    "/preferences/icon",
    { method: "POST", body: form },
    "Could not connect to the server to upload this icon. Check your connection and try again.",
  );
  await requireSuccessfulResponse(
    response,
    "The server could not upload this icon. Please try again.",
  );
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
  const response = await apiFetch(
    "/preferences/icon",
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ device_id: deviceId }),
    },
    "Could not connect to the server to remove this icon. The existing icon has been restored.",
  );
  await requireSuccessfulResponse(
    response,
    "The server could not remove this icon. The existing icon has been restored.",
  );
}

async function apiFetch(
  path: string,
  options: RequestInit,
  networkMessage: string,
): Promise<Response> {
  try {
    return await fetch(`${env.apiBaseUrl}${path}`, {
      ...options,
      credentials: "omit",
    });
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === "AbortError")
      throw cause;
    throw new Error(networkMessage, { cause });
  }
}

async function requireSuccessfulResponse(
  response: Response,
  fallbackMessage: string,
): Promise<void> {
  if (response.ok) return;
  const detail = (await response.text()).trim();
  if (detail && detail.length <= 160 && !detail.startsWith("<")) {
    throw new Error(detail.charAt(0).toUpperCase() + detail.slice(1));
  }
  throw new Error(fallbackMessage);
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
