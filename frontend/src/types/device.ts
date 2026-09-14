export interface Device {
  device_id: string;
  display_name: string;
  model: string;
  online: boolean;
  active_state: string;
  lat: number | null;
  lng: number | null;
  speed: number;
  last_reported_at: string | null;
}
