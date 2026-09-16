<script setup lang="ts">
import { computed, ref } from "vue";
import { useQuery, useQueryClient } from "@tanstack/vue-query";
import AppHeader from "@/components/AppHeader.vue";
import DeviceSidebar from "@/components/DeviceSidebar.vue";
import FleetMapPanel from "@/components/FleetMapPanel.vue";
import {
  fetchDevices,
  removeDeviceIcon,
  updateDeviceOrder,
  updateDevicePreferences,
  uploadDeviceIcon,
} from "@/lib/devicesApi";
import { reorderVisibleDevices } from "@/lib/deviceOrder";
import { createSerialTaskQueue } from "@/lib/serialTaskQueue";
import type { Device, DevicePreferenceUpdate } from "@/types/device";

const queryClient = useQueryClient();
const queryKey = ["devices"] as const;
const deviceRefreshInterval = 30_000;
const pendingWrites = ref(0);
const enqueueWrite = createSerialTaskQueue();
const { data, isPending, error, refetch } = useQuery({
  queryKey,
  queryFn: ({ signal }) => fetchDevices(signal),
  retry: false,
  staleTime: deviceRefreshInterval,
  refetchInterval: computed(() =>
    pendingWrites.value === 0 ? deviceRefreshInterval : false,
  ),
  refetchIntervalInBackground: false,
  refetchOnWindowFocus: computed(() =>
    pendingWrites.value === 0 ? "always" : false,
  ),
  refetchOnReconnect: computed(() =>
    pendingWrites.value === 0 ? "always" : false,
  ),
});
const devices = computed(() => data.value ?? []);
const loadError = computed(() => (data.value ? null : error.value));
const refreshError = computed(() =>
  data.value && error.value
    ? "Live updates are temporarily unavailable. Showing last known positions."
    : null,
);
const mapDevices = computed(() =>
  devices.value.filter((device) => !device.hidden),
);
const sidebarOpen = ref(true);
const selectedId = ref<string | null>(null);
const selectionRequest = ref(0);
const preferenceError = ref<string | null>(null);
const savingIds = ref(new Set<string>());

function selectDevice(deviceId: string) {
  selectedId.value = deviceId;
  sidebarOpen.value = true;
  selectionRequest.value += 1;
}

function retryDevices() {
  void refetch();
}

function withPausedPolling(action: () => Promise<void>) {
  pendingWrites.value++;
  void enqueueWrite(async () => {
    try {
      await queryClient.cancelQueries({ queryKey });
      await action();
    } catch (cause) {
      preferenceError.value = actionError(cause, "Could not save preferences");
    } finally {
      pendingWrites.value--;
      if (pendingWrites.value === 0) {
        void queryClient.invalidateQueries({ queryKey });
      }
    }
  });
}

function actionError(cause: unknown, fallback: string): string {
  return cause instanceof Error ? cause.message : fallback;
}

function preferenceFor(device: Device): DevicePreferenceUpdate {
  return {
    device_id: device.device_id,
    hidden: device.hidden,
    custom_display_name: device.custom_display_name,
  };
}

async function saveDevice(deviceId: string, changes: Partial<Device>) {
  const previous = queryClient.getQueryData<Device[]>(queryKey) ?? [];
  const current = previous.find((device) => device.device_id === deviceId);
  if (!current) return;
  const updated = { ...current, ...changes };
  if ("custom_display_name" in changes) {
    updated.display_name =
      changes.custom_display_name?.trim() || updated.original_display_name;
  }
  queryClient.setQueryData<Device[]>(queryKey, (devices = []) =>
    devices.map((device) => (device.device_id === deviceId ? updated : device)),
  );
  if (updated.hidden && selectedId.value === deviceId) {
    selectedId.value = null;
  }
  savingIds.value = new Set(savingIds.value).add(deviceId);
  preferenceError.value = null;
  try {
    await updateDevicePreferences(preferenceFor(updated));
  } catch (cause) {
    queryClient.setQueryData(queryKey, previous);
    preferenceError.value = actionError(cause, "Could not save preferences");
  } finally {
    const next = new Set(savingIds.value);
    next.delete(deviceId);
    savingIds.value = next;
  }
}

async function reorderDevices(deviceIds: string[]) {
  const previous = queryClient.getQueryData<Device[]>(queryKey) ?? [];
  const reordered = reorderVisibleDevices(previous, deviceIds);
  if (!reordered) return;
  queryClient.setQueryData(queryKey, reordered);
  preferenceError.value = null;
  try {
    await updateDeviceOrder(reordered.map((device) => device.device_id));
  } catch (cause) {
    queryClient.setQueryData(queryKey, previous);
    preferenceError.value = actionError(cause, "Could not save device order");
  }
}

async function uploadIcon(deviceId: string, file: File) {
  const previous = queryClient.getQueryData<Device[]>(queryKey) ?? [];
  savingIds.value = new Set(savingIds.value).add(deviceId);
  preferenceError.value = null;
  try {
    const uploaded = await uploadDeviceIcon(deviceId, file);
    queryClient.setQueryData<Device[]>(queryKey, (devices = []) =>
      devices.map((device) =>
        device.device_id === deviceId
          ? {
              ...device,
              icon_storage_path: uploaded.iconStoragePath,
              icon_url: uploaded.iconUrl,
            }
          : device,
      ),
    );
  } catch (cause) {
    queryClient.setQueryData(queryKey, previous);
    preferenceError.value = actionError(cause, "Could not upload icon");
  } finally {
    const next = new Set(savingIds.value);
    next.delete(deviceId);
    savingIds.value = next;
  }
}

async function removeIcon(deviceId: string) {
  const previous = queryClient.getQueryData<Device[]>(queryKey) ?? [];
  queryClient.setQueryData<Device[]>(queryKey, (devices = []) =>
    devices.map((device) =>
      device.device_id === deviceId
        ? { ...device, icon_storage_path: null, icon_url: null }
        : device,
    ),
  );
  savingIds.value = new Set(savingIds.value).add(deviceId);
  preferenceError.value = null;
  try {
    await removeDeviceIcon(deviceId);
  } catch (cause) {
    queryClient.setQueryData(queryKey, previous);
    preferenceError.value = actionError(cause, "Could not remove icon");
  } finally {
    const next = new Set(savingIds.value);
    next.delete(deviceId);
    savingIds.value = next;
  }
}
</script>

<template>
  <div class="app-shell">
    <AppHeader />
    <main
      class="workspace"
      :class="{ 'sidebar-collapsed': !sidebarOpen }"
    >
      <DeviceSidebar
        :devices="devices"
        :selected-id="selectedId"
        :open="sidebarOpen"
        :selection-request="selectionRequest"
        :loading="isPending"
        :error="loadError?.message ?? null"
        :preference-error="preferenceError ?? refreshError"
        :saving-ids="savingIds"
        :busy="pendingWrites > 0"
        @select="selectDevice"
        @close="sidebarOpen = false"
        @retry="retryDevices"
        @update-device="
          (id, changes) => withPausedPolling(() => saveDevice(id, changes))
        "
        @reorder="(ids) => withPausedPolling(() => reorderDevices(ids))"
        @upload-icon="
          (id, file) => withPausedPolling(() => uploadIcon(id, file))
        "
        @remove-icon="(id) => withPausedPolling(() => removeIcon(id))"
      />
      <FleetMapPanel
        :devices="mapDevices"
        :selected-id="selectedId"
        :sidebar-open="sidebarOpen"
        :loading="isPending"
        :error="!!loadError"
        @select="selectDevice"
        @open-sidebar="sidebarOpen = true"
      />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  background: #edf1f4;
  color: #213248;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
}
.workspace {
  position: relative;
  height: calc(100vh - 72px);
  min-height: 570px;
  overflow: hidden;
}
.workspace > :first-child {
  position: absolute;
  z-index: 10;
  top: 18px;
  bottom: 18px;
  left: 18px;
  width: min(calc(33.333% - 18px), 420px);
  transition:
    transform 0.42s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.25s ease;
}
.workspace.sidebar-collapsed > :first-child {
  transform: translateX(calc(-100% - 20px));
  opacity: 0;
}
.workspace > :last-child {
  width: 100%;
  height: 100%;
}
@media (max-width: 760px) {
  .workspace {
    height: calc(100dvh - 62px);
    min-height: 500px;
  }
  .workspace > :first-child {
    top: auto;
    bottom: 12px;
    left: 12px;
    width: calc(100% - 24px);
    height: min(68%, 620px);
  }
  .workspace.sidebar-collapsed > :first-child {
    transform: translateY(calc(100% + 16px));
  }
}
@media (prefers-reduced-motion: reduce) {
  .workspace > :first-child {
    transition: none;
  }
}
</style>
