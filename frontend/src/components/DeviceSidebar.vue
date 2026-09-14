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
import DeviceListControls from "./DeviceListControls.vue";
import SummaryCard from "./SummaryCard.vue";
import type { DeviceFilter } from "./DeviceListControls.vue";
import type { Device } from "@/types/device";

const props = defineProps<{
  devices: Device[];
  selectedId: string | null;
  open: boolean;
  selectionRequest: number;
}>();
const emit = defineEmits<{ select: [deviceId: string]; close: [] }>();

const search = ref("");
const filter = ref<DeviceFilter>("all");
const scrollContainer = ref<HTMLElement | null>(null);
const listElement = ref<HTMLElement | null>(null);
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
    return (
      (!query ||
        `${device.display_name} ${device.model}`
          .toLowerCase()
          .includes(query)) &&
      (filter.value === "all" ||
        (filter.value === "online" ? device.online : !device.online))
    );
  }),
);

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

      <div class="summary-grid">
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
      </div>

      <div class="list-heading">
        <strong
          >Your devices <span>{{ devices.length }}</span></strong
        >
      </div>
      <DeviceListControls v-model:search="search" v-model:filter="filter" />
      <div ref="listElement" class="device-list">
        <DeviceCard
          v-for="device in filteredDevices"
          :key="device.device_id"
          :device="device"
          :selected="selectedId === device.device_id"
          :data-device-id="device.device_id"
          @select="emit('select', $event)"
        />
        <div v-if="filteredDevices.length === 0" class="empty-list">
          <Search :size="21" /><strong>No devices found</strong
          ><span>Try another search or filter.</span>
        </div>
      </div>
    </div>
    <div class="sidebar-footer"><i /> Showing sample device data</div>
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
