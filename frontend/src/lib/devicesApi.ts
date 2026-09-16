import { z } from "zod";
import { env } from "@/lib/env";
import type { Device } from "@/types/device";

const deviceResponseSchema = z.object({
  result_list: z.array(
    z.object({
      device_id: z.string(),
      display_name: z.string(),
      model: z.string(),
      online: z.boolean(),
      active_state: z.string(),
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
    return {
      device_id: device.device_id,
      display_name: device.display_name,
      model: device.model,
      online: device.online,
      active_state: device.active_state,
      lat: point?.lat ?? null,
      lng: point?.lng ?? null,
      // The upstream point speed is in km/h; cards display mph.
      speed: (point?.speed ?? 0) / 1.609344,
      last_reported_at: point?.dt_tracker ?? point?.dt_server ?? null,
    };
  });
}
