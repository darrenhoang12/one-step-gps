import type { Device } from "@/types/device";

export function formatReportedAt(value: string | null): string {
  if (!value) return "No recent report";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "No recent report";
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

export function formatCoordinates(device: Device): string {
  if (device.lat === null || device.lng === null) return "Location unavailable";
  return `${device.lat.toFixed(4)}° N, ${Math.abs(device.lng).toFixed(4)}° W`;
}
