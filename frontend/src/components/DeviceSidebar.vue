<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import {
  Activity,
  ChevronLeft,
  Gauge,
  LayoutGrid,
  Search,
  Truck,
} from "@lucide/vue";
import DeviceCard from "./DeviceCard.vue";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import DeviceListControls from "./DeviceListControls.vue";
import SummaryCard from "./SummaryCard.vue";
import type { DeviceFilter } from "./DeviceListControls.vue";
import type { Device } from "@/types/device";

const props = defineProps<{
  devices: Device[];
  selectedId: string | null;
  open: boolean;
  selectionRequest: number;
  loading: boolean;
  error: string | null;
  preferenceError: string | null;
  savingIds: Set<string>;
}>();
const emit = defineEmits<{
  select: [deviceId: string];
  close: [];
  retry: [];
  updateDevice: [deviceId: string, changes: Partial<Device>];
  reorder: [deviceIds: string[]];
  uploadIcon: [deviceId: string, file: File];
}>();

const search = ref("");
const filter = ref<DeviceFilter>("all");
const scrollContainer = ref<HTMLElement | null>(null);
const listElement = ref<HTMLElement | null>(null);
const draggedId = ref<string | null>(null);
const onlineCount = computed(
  () => props.devices.filter((device) => device.online).length,
);
const movingCount = computed(
  () =>
    props.devices.filter((device) => device.online && device.speed > 2).length,
);
const filteredDevices = computed(() =>
  props.devices.filter((device) => {
    const query = search.value.trim().toLowerCase();
    const matchesState =
      filter.value === "hidden"
        ? device.hidden && !device.archived
        : filter.value === "archived"
          ? device.archived
          : !device.hidden && !device.archived &&
            (filter.value === "all" ||
              (filter.value === "online" ? device.online : !device.online));
    return (
      (!query ||
        `${device.display_name} ${device.original_display_name} ${device.model}`
          .toLowerCase()
          .includes(query)) &&
      matchesState
    );
  }),
);
const canReorder = computed(() => filter.value === "all" && !search.value.trim());

function dropDevice(targetId: string) {
  const sourceId = draggedId.value;
  draggedId.value = null;
  if (!sourceId || sourceId === targetId || !canReorder.value) return;
  const ordered = filteredDevices.value.map((device) => device.device_id);
  const sourceIndex = ordered.indexOf(sourceId);
  const targetIndex = ordered.indexOf(targetId);
  if (sourceIndex < 0 || targetIndex < 0) return;
  ordered.splice(targetIndex, 0, ...ordered.splice(sourceIndex, 1));
  emit("reorder", ordered);
}

function updateDevice(deviceId: string, changes: Partial<Device>) {
  emit("updateDevice", deviceId, changes);
}

function uploadIcon(deviceId: string, file: File) {
  emit("uploadIcon", deviceId, file);
}

watch(
  () => props.selectionRequest,
  async () => {
    if (!props.selectedId) return;
    // A map selection may be hidden by the current search or status filter.
    if (
      !filteredDevices.value.some(
        (device) => device.device_id === props.selectedId,
      )
    ) {
      search.value = "";
      filter.value = "all";
    }
    await nextTick();
    const container = scrollContainer.value;
    const card = Array.from(
      listElement.value?.querySelectorAll<HTMLElement>("[data-device-id]") ??
        [],
    ).find((element) => element.dataset.deviceId === props.selectedId);
    if (!container || !card) return;
    const top =
      card.getBoundingClientRect().top -
      container.getBoundingClientRect().top +
      container.scrollTop -
      20;
    container.scrollTo({
      top,
      behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches
        ? "instant"
        : "smooth",
    });
  },
);
</script>

<template>
  <aside
    class="sidebar"
    :class="{ open }"
    :inert="!open"
    aria-label="Devices panel"
  >
    <div ref="scrollContainer" class="sidebar-scroll">
      <div class="sidebar-heading">
        <div>
          <div class="eyebrow"><LayoutGrid :size="12" /> FLEET OVERVIEW</div>
          <h1>Devices</h1>
        </div>
        <Button
          variant="outline"
          size="icon"
          class="icon-button"
          aria-label="Collapse devices panel"
          title="Collapse devices panel"
          @click="emit('close')"
        >
          <ChevronLeft :size="20" />
        </Button>
      </div>

      <div v-if="loading || !error" class="summary-grid">
        <template v-if="loading">
          <div v-for="index in 3" :key="index" class="summary-placeholder">
            <Skeleton class="h-6 w-6 rounded-md" />
            <Skeleton class="mt-2 h-5 w-10" />
            <Skeleton class="mt-1 h-3 w-16" />
          </div>
        </template>
        <template v-else>
          <SummaryCard
            :value="devices.length"
            label="Total devices"
            :icon="Truck"
            tone="blue"
          />
          <SummaryCard
            :value="onlineCount"
            label="Online now"
            :icon="Activity"
            tone="green"
          />
          <SummaryCard
            :value="movingCount"
            label="Moving"
            :icon="Gauge"
            tone="orange"
          />
        </template>
      </div>

      <div class="list-heading">
        <strong>
          Your devices
          <span v-if="!loading && !error">{{ devices.length }}</span>
        </strong>
      </div>
      <div v-if="preferenceError" class="preference-error" role="alert">
        {{ preferenceError }}
      </div>
      <template v-if="loading">
        <Skeleton class="mt-4 h-10 w-full rounded-md" />
        <Skeleton class="my-4 h-8 w-48 rounded-md" />
        <div class="device-list" aria-label="Loading devices" role="status">
          <div v-for="index in 4" :key="index" class="device-placeholder">
            <Skeleton class="h-9 w-9 rounded-md" />
            <div class="placeholder-lines">
              <Skeleton class="h-4 w-32" />
              <Skeleton class="mt-2 h-3 w-24" />
            </div>
            <Skeleton class="mt-4 h-3 w-36" />
          </div>
        </div>
      </template>
      <div v-else-if="error" class="load-error" role="alert">
        <strong>Could not load devices</strong>
        <span>{{ error }}</span>
        <Button variant="outline" size="sm" @click="emit('retry')"
          >Try again</Button
        >
      </div>
      <template v-else>
        <DeviceListControls v-model:search="search" v-model:filter="filter" />
        <div ref="listElement" class="device-list">
          <DeviceCard
            v-for="device in filteredDevices"
            :key="device.device_id"
            :device="device"
            :selected="selectedId === device.device_id"
            :data-device-id="device.device_id"
            :draggable="canReorder"
            :saving="savingIds.has(device.device_id)"
            @select="emit('select', $event)"
            @update="updateDevice"
            @upload="uploadIcon"
            @drag-start="draggedId = $event"
            @drop="dropDevice"
          />
          <div v-if="filteredDevices.length === 0" class="empty-list">
            <Search :size="21" /><strong>No devices found</strong
            ><span>{{
              devices.length
                ? "Try another search or filter."
                : "No devices are available."
            }}</span>
          </div>
        </div>
      </template>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  min-width: 0;
  overflow: hidden;
  background: #f7f9fb;
  border: 1px solid #d9e0e7;
  border-radius: 18px;
  box-shadow:
    0 18px 55px #152b3d30,
    0 2px 8px #152b3d12;
  display: flex;
  flex-direction: column;
}
.sidebar-scroll {
  padding: 30px 24px 20px;
  overflow-y: auto;
  min-height: 0;
  scrollbar-color: #c6d1dc transparent;
  scrollbar-width: thin;
}
.sidebar-scroll,
.sidebar-footer {
  transition:
    opacity 0.2s ease,
    transform 0.38s cubic-bezier(0.22, 1, 0.36, 1);
}
.sidebar:not(.open) .sidebar-scroll,
.sidebar:not(.open) .sidebar-footer {
  opacity: 0;
  transform: translateX(-12px);
}
.sidebar-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.eyebrow {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  color: #e56843;
  font-weight: 800;
  letter-spacing: 0.15em;
}
.sidebar-heading h1 {
  margin: 11px 0 3px;
  font-size: 32px;
  line-height: 1.1;
  letter-spacing: -0.055em;
  font-weight: 800;
}
.sidebar-heading p {
  margin: 0;
  color: #8190a0;
  font-size: 13px;
}
.icon-button {
  display: grid;
  place-items: center;
  flex: none;
  width: 34px;
  height: 34px;
  border: 1px solid #dae2e9;
  border-radius: 9px;
  background: #fff;
  color: #64758a;
  cursor: pointer;
}
.icon-button:hover {
  color: #1a385e;
  background: #edf3f9;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 9px;
  margin-top: 27px;
}
.summary-placeholder,
.device-placeholder {
  padding: 13px;
  border: 1px solid #e6ebf0;
  border-radius: 12px;
  background: #fff;
}
.device-placeholder {
  min-height: 112px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex-wrap: wrap;
}
.placeholder-lines {
  flex: 1;
}
.load-error {
  display: flex;
  align-items: flex-start;
  flex-direction: column;
  gap: 10px;
  margin-top: 18px;
  padding: 18px;
  border: 1px solid #e6ebf0;
  border-radius: 12px;
  background: #fff;
  color: #64758a;
  font-size: 12px;
}
.load-error strong {
  color: #213248;
  font-size: 14px;
}
.preference-error {
  margin-top: 12px;
  padding: 9px 11px;
  border: 1px solid #efc1b5;
  border-radius: 8px;
  background: #fff3ef;
  color: #a44a31;
  font-size: 11px;
}
.list-heading {
  display: flex;
  flex-direction: column;
  gap: 3px;
  margin-top: 29px;
}
.list-heading strong {
  font-size: 15px;
  font-weight: 800;
}
.list-heading strong span {
  margin-left: 4px;
  padding: 2px 7px;
  border-radius: 20px;
  background: #e7edf4;
  font-size: 11px;
  color: #6d7d8f;
}
.list-heading small {
  color: #8997a6;
  font-size: 11px;
}
.device-list {
  display: grid;
  gap: 10px;
}
.empty-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 30px;
  border: 1px dashed #d3dce5;
  border-radius: 12px;
  color: #9caab6;
  font-size: 11px;
}
.empty-list strong {
  color: #60758a;
  font-size: 13px;
}
.sidebar-footer {
  flex: none;
  display: flex;
  align-items: center;
  gap: 7px;
  min-height: 44px;
  padding: 0 24px;
  border-top: 1px solid #e6ebf0;
  color: #9aa7b4;
  font-size: 10px;
}
.sidebar-footer i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #c1cfdb;
}
@media (max-width: 760px) {
  .sidebar {
    border-radius: 18px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .sidebar,
  .sidebar-scroll,
  .sidebar-footer {
    transition: none;
  }
}
</style>
