<script setup lang="ts">
import { computed, ref } from "vue";
import { ChevronRight, Crosshair, MapPin, Menu } from "@lucide/vue";
import DeviceMap from "./DeviceMap.vue";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import SelectedDeviceCard from "./SelectedDeviceCard.vue";
import type { Device } from "@/types/device";

const props = defineProps<{
  devices: Device[];
  selectedId: string | null;
  sidebarOpen: boolean;
  loading: boolean;
  error: boolean;
}>();
const emit = defineEmits<{ select: [deviceId: string]; openSidebar: [] }>();
const mapRef = ref<InstanceType<typeof DeviceMap> | null>(null);
const selectedDevice = computed(
  () =>
    props.devices.find((device) => device.device_id === props.selectedId) ??
    null,
);
</script>

<template>
  <section
    class="map-panel"
    :class="{ 'sidebar-open': sidebarOpen }"
    aria-label="Device map"
  >
    <DeviceMap
      v-if="!loading && !error"
      ref="mapRef"
      :devices="devices"
      :selected-id="selectedId"
      :sidebar-open="sidebarOpen"
      @select="emit('select', $event)"
    />
    <div v-else class="map-loading" role="status" :aria-label="loading ? 'Loading map' : 'Map unavailable'">
      <Skeleton v-if="loading" class="h-full w-full rounded-none" />
      <span v-else>Map unavailable while devices could not be loaded.</span>
    </div>
    <div class="map-top-overlay">
      <Button
        v-if="!sidebarOpen"
        variant="outline"
        class="reopen-button"
        aria-label="Open devices panel"
        @click="emit('openSidebar')"
      >
        <Menu :size="18" /><span>Devices</span><ChevronRight :size="16" />
      </Button>
      <div class="map-heading">
        <span class="map-heading-icon"><MapPin :size="18" /></span>
        <div>
          <strong>Fleet map</strong
          ><small>{{ loading ? "Loading devices" : error ? "Devices unavailable" : `${devices.length} devices` }}</small>
        </div>
      </div>
    </div>
    <div v-if="selectedDevice" class="selected-device-overlay">
      <SelectedDeviceCard :device="selectedDevice" />
    </div>
    <div class="map-bottom-overlay">
      <Button
        variant="outline"
        class="map-action"
        title="Fit all devices"
        aria-label="Fit all devices on map"
        :disabled="loading || error"
        @click="mapRef?.fitAll()"
      >
        <Crosshair :size="17" /><span>Fit all</span>
      </Button>
      <div class="map-legend">
        <i class="legend-online" /> Online <i class="legend-offline" /> Offline
      </div>
    </div>
  </section>
</template>

<style scoped>
.map-panel {
  position: relative;
  min-width: 0;
  overflow: hidden;
  background: #e5e8dc;
}
.map-loading {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: #607487;
  font-size: 13px;
}
.map-top-overlay,
.map-bottom-overlay {
  position: absolute;
  z-index: 5;
  left: 23px;
  right: 23px;
  pointer-events: none;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}
.map-top-overlay {
  top: 22px;
}
.selected-device-overlay {
  position: absolute;
  z-index: 5;
  top: 22px;
  right: 23px;
  width: min(466px, calc(100% - 46px));
  pointer-events: none;
}
.selected-device-overlay > * {
  pointer-events: auto;
}
.map-panel.sidebar-open .map-top-overlay,
.map-panel.sidebar-open .map-bottom-overlay {
  left: min(calc(33.333% + 18px), 456px);
}
.map-bottom-overlay {
  bottom: 23px;
  align-items: center;
  justify-content: flex-end;
}
.map-top-overlay > *,
.map-bottom-overlay > * {
  pointer-events: auto;
}
.map-heading {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 14px 11px 11px;
  border: 1px solid #ffffffbb;
  border-radius: 12px;
  background: #fffcf6f0;
  box-shadow: 0 4px 18px #30435923;
  backdrop-filter: blur(10px);
}
.map-heading-icon {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #e8eff5;
  color: #315f88;
}
.map-heading strong,
.map-heading small {
  display: block;
  white-space: nowrap;
}
.map-heading strong {
  font-size: 12px;
}
.map-heading small {
  margin-top: 2px;
  color: #8b9aa6;
  font-size: 10px;
}
.reopen-button,
.map-action {
  display: flex;
  align-items: center;
  gap: 7px;
  height: 52px;
  padding: 0 13px;
  border: 1px solid #ffffffbb;
  border-radius: 12px;
  background: #fffcf6f0;
  box-shadow: 0 4px 18px #30435923;
  backdrop-filter: blur(10px);
  color: #34506c;
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
}
.reopen-button:hover,
.map-action:hover {
  background: #fff;
}
.map-action {
  height: 36px;
}
.map-legend {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 9px 12px;
  border-radius: 8px;
  background: #fffffff0;
  box-shadow: 0 4px 15px #1a355020;
  color: #607487;
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
}
.legend-online,
.legend-offline {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #28557e;
}
.legend-offline {
  margin-left: 5px;
  background: #9aa5ac;
}
@media (min-width: 761px) and (max-width: 1100px) {
  .map-panel.sidebar-open .selected-device-overlay {
    top: 88px;
  }
}
@media (max-width: 760px) {
  .map-panel {
    height: 100%;
  }
  .map-top-overlay {
    left: 14px;
    right: 14px;
    top: 14px;
  }
  .map-bottom-overlay {
    left: 14px;
    right: 14px;
    bottom: 14px;
  }
  .selected-device-overlay {
    top: 78px;
    right: 14px;
    width: min(466px, calc(100% - 28px));
  }
  .map-panel.sidebar-open .selected-device-overlay {
    display: none;
  }
  .map-panel.sidebar-open .map-top-overlay,
  .map-panel.sidebar-open .map-bottom-overlay {
    left: 14px;
  }
  .map-panel.sidebar-open .map-bottom-overlay {
    display: none;
  }
  .map-action span {
    display: none;
  }
  .map-action {
    width: 36px;
    justify-content: center;
  }
  .reopen-button {
    padding: 0 10px;
  }
  .reopen-button span {
    display: none;
  }
}
@media (max-width: 420px) {
  .map-heading small {
    display: none;
  }
  .map-heading {
    height: 52px;
  }
}
</style>
