<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { env } from "@/lib/env";
import type { Device } from "@/types/device";

type LatLng = { lat: number; lng: number };
type MapInstance = {
  fitBounds: (
    bounds: BoundsInstance,
    padding?:
      number | { left: number; right: number; top: number; bottom: number },
  ) => void;
  panTo: (position: LatLng) => void;
  panBy: (x: number, y: number) => void;
  setZoom: (zoom: number) => void;
};
type BoundsInstance = { extend: (position: LatLng) => void };
type MarkerInstance = {
  addListener: (event: string, callback: () => void) => void;
};
type MapsApi = {
  maps: {
    Map: new (
      element: HTMLElement,
      options: Record<string, unknown>,
    ) => MapInstance;
    LatLngBounds: new () => BoundsInstance;
    marker: {
      AdvancedMarkerElement: new (options: {
        map: MapInstance;
        position: LatLng;
        title: string;
        content: HTMLElement;
      }) => MarkerInstance;
    };
  };
};

declare global {
  interface Window {
    google?: MapsApi;
    __oneStepMapsReady?: () => void;
  }
}

const props = defineProps<{
  devices: Device[];
  selectedId: string | null;
  sidebarOpen: boolean;
}>();
const emit = defineEmits<{ select: [deviceId: string] }>();

const apiKey = env.googleMapsApiKey ?? "";
const mapState = ref<"preview" | "loading" | "ready" | "error">(
  apiKey ? "loading" : "preview",
);
const mapElement = ref<HTMLElement | null>(null);
const markerElements = new Map<string, HTMLElement>();
let map: MapInstance | null = null;
let initialBounds: BoundsInstance | null = null;
let mapsPromise: Promise<MapsApi> | null = null;

const locatedDevices = computed(() =>
  props.devices.filter(
    (device): device is Device & LatLng =>
      typeof device.lat === "number" && typeof device.lng === "number",
  ),
);

function loadMaps(key: string): Promise<MapsApi> {
  if (window.google?.maps?.Map) return Promise.resolve(window.google);
  if (mapsPromise) return mapsPromise;

  mapsPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    const params = new URLSearchParams({
      key,
      v: "weekly",
      libraries: "marker",
      loading: "async",
      callback: "__oneStepMapsReady",
    });

    window.__oneStepMapsReady = () => {
      if (window.google) resolve(window.google);
      else reject(new Error("Google Maps did not initialize"));
      delete window.__oneStepMapsReady;
    };
    script.onerror = () => {
      mapsPromise = null;
      delete window.__oneStepMapsReady;
      reject(new Error("Could not load Google Maps"));
    };
    script.src = `https://maps.googleapis.com/maps/api/js?${params.toString()}`;
    script.async = true;
    document.head.appendChild(script);
  });

  return mapsPromise;
}

function fitAll() {
  if (map && initialBounds)
    map.fitBounds(initialBounds, {
      left:
        props.sidebarOpen && window.innerWidth > 760
          ? Math.min(window.innerWidth / 3, 420) + 70
          : 70,
      right: 70,
      top: 80,
      bottom: 80,
    });
}

function focusDevice(deviceId: string | null) {
  if (!map || !deviceId) return;
  const device = locatedDevices.value.find(
    (item) => item.device_id === deviceId,
  );
  if (!device) return;
  map.panTo({ lat: device.lat, lng: device.lng });
  map.setZoom(12);
  if (props.sidebarOpen && window.innerWidth > 760) {
    map.panBy((Math.min(window.innerWidth / 3, 420) + 18) / 2, 0);
  }
}

function previewPosition(device: Device & LatLng) {
  const longitudes = locatedDevices.value.map((item) => item.lng);
  const latitudes = locatedDevices.value.map((item) => item.lat);
  const west = Math.min(...longitudes) - 0.4;
  const east = Math.max(...longitudes) + 0.4;
  const south = Math.min(...latitudes) - 0.4;
  const north = Math.max(...latitudes) + 0.4;
  const horizontalPosition = 12 + ((device.lng - west) / (east - west)) * 76;
  const panelPercent =
    props.sidebarOpen && window.innerWidth > 760
      ? (Math.min(window.innerWidth / 3, 420) / window.innerWidth) * 100 + 2
      : 0;
  return {
    left: `${panelPercent + (horizontalPosition * (100 - panelPercent)) / 100}%`,
    top: `${12 + ((north - device.lat) / (north - south)) * 76}%`,
  };
}

onMounted(async () => {
  if (!apiKey || !mapElement.value) return;

  try {
    const google = await loadMaps(apiKey);
    if (!mapElement.value) return;

    map = new google.maps.Map(mapElement.value, {
      center: { lat: 36.1, lng: -119.7 },
      zoom: 6,
      mapId: env.googleMapsMapId,
      mapTypeControl: false,
      streetViewControl: false,
      fullscreenControl: false,
      gestureHandling: "greedy",
    });

    const bounds = new google.maps.LatLngBounds();
    for (const device of locatedDevices.value) {
      const position = { lat: device.lat, lng: device.lng };
      const pin = document.createElement("div");
      pin.className = `google-device-pin ${device.online ? "is-online" : "is-offline"}`;
      pin.classList.toggle(
        "is-selected",
        device.device_id === props.selectedId,
      );
      pin.textContent = device.display_name.slice(0, 2).toUpperCase();
      pin.setAttribute("aria-label", device.display_name);
      markerElements.set(device.device_id, pin);

      const marker = new google.maps.marker.AdvancedMarkerElement({
        map,
        position,
        title: device.display_name,
        content: pin,
      });
      marker.addListener("click", () => emit("select", device.device_id));
      bounds.extend(position);
    }

    initialBounds = bounds;
    if (locatedDevices.value.length > 0) fitAll();
    mapState.value = "ready";
  } catch (error) {
    console.error("Google Maps failed to load:", error);
    mapState.value = "error";
  }
});

watch(
  () => props.selectedId,
  (deviceId) => {
    for (const [id, element] of markerElements) {
      element.classList.toggle("is-selected", id === deviceId);
    }
    focusDevice(deviceId);
  },
);

watch(
  () => props.sidebarOpen,
  () => focusDevice(props.selectedId),
);

defineExpose({ fitAll });
</script>

<template>
  <div class="map-surface">
    <div
      ref="mapElement"
      class="google-map"
      :class="{ visible: mapState === 'ready' }"
    />

    <div v-if="mapState !== 'ready'" class="map-preview">
      <div class="terrain terrain-one" />
      <div class="terrain terrain-two" />
      <div class="terrain terrain-three" />
      <div class="road road-one" />
      <div class="road road-two" />
      <div class="road road-three" />
      <div class="road road-four" />
      <span class="map-region region-la">LOS ANGELES</span>
      <span class="map-region region-bay">BAY AREA</span>
      <span class="map-region region-sac">SACRAMENTO</span>
      <span class="map-region region-ocean">PACIFIC OCEAN</span>

      <button
        v-for="device in locatedDevices"
        :key="device.device_id"
        class="preview-pin"
        :class="{
          selected: device.device_id === selectedId,
          offline: !device.online,
        }"
        :style="previewPosition(device)"
        :title="device.display_name"
        :aria-label="`Select ${device.display_name}`"
        @click="emit('select', device.device_id)"
      >
        <span class="pin-pulse" />
        <span class="pin-core">{{
          device.display_name.slice(0, 2).toUpperCase()
        }}</span>
      </button>

      <div class="map-preview-notice">
        <span class="notice-dot" />
        <div>
          <strong>{{
            mapState === "loading" ? "Loading Google Maps" : "Map preview"
          }}</strong>
          <p>
            {{
              mapState === "error"
                ? "Google Maps could not load. Check your API key and billing settings."
                : apiKey
                  ? "Preparing the live map and device markers."
                  : "Add VITE_GOOGLE_MAPS_API_KEY to show the live Google map."
            }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-surface,
.google-map,
.map-preview {
  position: absolute;
  inset: 0;
}
.google-map {
  opacity: 0;
  transition: opacity 0.3s ease;
}
.google-map.visible {
  opacity: 1;
  z-index: 1;
}
.map-preview {
  overflow: hidden;
  background: #e7e8db;
  background-image:
    linear-gradient(
      27deg,
      transparent 48%,
      #f5f2e7 49%,
      #f5f2e7 51%,
      transparent 52%
    ),
    linear-gradient(
      152deg,
      transparent 47%,
      #d6ddce 48%,
      #d6ddce 52%,
      transparent 53%
    );
  background-size:
    210px 180px,
    300px 260px;
}
.terrain {
  position: absolute;
  border-radius: 50%;
  filter: blur(2px);
}
.terrain-one {
  width: 85%;
  height: 55%;
  background: #d5dbc4;
  transform: rotate(-18deg);
  left: 14%;
  top: -11%;
}
.terrain-two {
  width: 70%;
  height: 45%;
  background: #dce3cf;
  transform: rotate(24deg);
  left: 38%;
  bottom: -10%;
}
.terrain-three {
  width: 50%;
  height: 35%;
  background: #c9d7cd;
  left: -28%;
  bottom: -12%;
}
.road {
  position: absolute;
  height: 8px;
  background: #fffaf0;
  border: 1px solid #d9d8c9;
  box-shadow: 0 0 0 5px #f4f0e2;
  border-radius: 99px;
  transform-origin: left center;
}
.road-one {
  width: 125%;
  left: -15%;
  top: 33%;
  transform: rotate(25deg);
}
.road-two {
  width: 115%;
  left: 10%;
  top: 81%;
  transform: rotate(-37deg);
}
.road-three {
  width: 110%;
  left: -8%;
  top: 60%;
  transform: rotate(-13deg);
}
.road-four {
  width: 80%;
  left: 36%;
  top: 13%;
  transform: rotate(76deg);
}
.map-region {
  position: absolute;
  z-index: 1;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.2em;
  color: #87948e;
  white-space: nowrap;
}
.region-la {
  left: 58%;
  top: 72%;
}
.region-bay {
  left: 19%;
  top: 24%;
}
.region-sac {
  left: 35%;
  top: 10%;
}
.region-ocean {
  left: 6%;
  bottom: 12%;
  color: #91aaa8;
  transform: rotate(-28deg);
}
.preview-pin {
  position: absolute;
  z-index: 3;
  border: 0;
  background: transparent;
  cursor: pointer;
  transform: translate(-50%, -50%);
  padding: 10px;
}
.pin-pulse {
  position: absolute;
  inset: 2px;
  border-radius: 50%;
  background: #173c68;
  opacity: 0.12;
  transition: transform 0.2s;
}
.preview-pin.selected .pin-pulse {
  transform: scale(1.5);
  opacity: 0.18;
}
.pin-core {
  position: relative;
  display: grid;
  place-items: center;
  width: 37px;
  height: 37px;
  border-radius: 50%;
  background: #173c68;
  border: 3px solid white;
  box-shadow: 0 5px 18px #132d4955;
  color: white;
  font-size: 10px;
  font-weight: 800;
}
.preview-pin.offline .pin-core {
  background: #8b9799;
}
.preview-pin.selected .pin-core {
  background: #e7653f;
  width: 45px;
  height: 45px;
  font-size: 11px;
}
.map-preview-notice {
  position: absolute;
  z-index: 4;
  left: 50%;
  bottom: 50px;
  transform: translateX(-50%);
  width: min(90%, 410px);
  display: flex;
  gap: 12px;
  padding: 14px 18px;
  border-radius: 14px;
  background: #fffefaee;
  box-shadow: 0 10px 38px #233d4a22;
  backdrop-filter: blur(12px);
  color: #21344a;
}
.notice-dot {
  flex: none;
  width: 9px;
  height: 9px;
  margin-top: 5px;
  border-radius: 50%;
  background: #e7653f;
  box-shadow: 0 0 0 4px #f8d9cf;
}
.map-preview-notice strong {
  font-size: 13px;
}
.map-preview-notice p {
  margin: 3px 0 0;
  font-size: 11px;
  line-height: 1.5;
  color: #657484;
}
:global(.google-device-pin) {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border: 3px solid white;
  border-radius: 50%;
  color: white;
  background: #173c68;
  box-shadow: 0 5px 18px #132d4955;
  font-size: 10px;
  font-weight: 800;
  cursor: pointer;
}
:global(.google-device-pin.is-offline) {
  background: #8b9799;
}
:global(.google-device-pin.is-selected) {
  background: #e7653f;
  width: 44px;
  height: 44px;
}
</style>
