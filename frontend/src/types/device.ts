export interface Device {
  device_id: string;
  display_name: string;
  original_display_name: string;
  model: string;
  online: boolean;
  active_state: string;
  lat: number | null;
  lng: number | null;
  speed: number;
  last_reported_at: string | null;
  sort_order: number | null;
  hidden: boolean;
  custom_display_name: string | null;
  icon_storage_path: string | null;
  icon_url: string | null;
}

export interface DevicePreferenceUpdate {
  device_id: string;
  sort_order: number | null;
  hidden: boolean;
  custom_display_name: string | null;
}
