import type { Device } from "@/types/device";

// Preserve hidden devices in their existing slots while moving visible cards.
export function reorderVisibleDevices(
  devices: Device[],
  visibleIds: string[],
): Device[] | null {
  const visible = devices.filter((device) => !device.hidden);
  if (visible.length !== visibleIds.length) return null;

  const remaining = new Set(visible.map((device) => device.device_id));
  for (const id of visibleIds) {
    if (!remaining.delete(id)) return null;
  }

  const byId = new Map(devices.map((device) => [device.device_id, device]));
  let nextVisible = 0;
  return devices.map((device, index) => {
    const id = device.hidden ? device.device_id : visibleIds[nextVisible++]!;
    return { ...byId.get(id)!, sort_order: index };
  });
}
