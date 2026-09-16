import type { Device } from "@/types/device";

export type DeviceMovement = "moving" | "stopped" | "unknown";

export function deviceMovement(device: Pick<Device, "online" | "speed">): DeviceMovement {
  if (!device.online) return "unknown";
  return device.speed > 2 ? "moving" : "stopped";
}
