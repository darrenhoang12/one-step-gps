<script setup lang="ts">
import { computed, ref } from "vue";
import { useQuery } from "@tanstack/vue-query";
import AppHeader from "@/components/AppHeader.vue";
import DeviceSidebar from "@/components/DeviceSidebar.vue";
import FleetMapPanel from "@/components/FleetMapPanel.vue";
import { fetchDevices } from "@/lib/devicesApi";

const { data, isPending, error, refetch } = useQuery({
  queryKey: ["devices"],
  queryFn: ({ signal }) => fetchDevices(signal),
  retry: false,
  staleTime: Infinity,
  refetchOnWindowFocus: false,
  refetchOnReconnect: false,
});
const devices = computed(() => data.value ?? []);
const sidebarOpen = ref(true);
const selectedId = ref<string | null>(null);
const selectionRequest = ref(0);

function selectDevice(deviceId: string) {
  selectedId.value = deviceId;
  sidebarOpen.value = true;
  selectionRequest.value += 1;
}

function retryDevices() {
  void refetch();
}
</script>

<template>
  <div class="app-shell">
    <AppHeader />
    <main class="workspace" :class="{ 'sidebar-collapsed': !sidebarOpen }">
      <DeviceSidebar
        :devices="devices"
        :selected-id="selectedId"
        :open="sidebarOpen"
        :selection-request="selectionRequest"
        :loading="isPending"
        :error="error?.message ?? null"
        @select="selectDevice"
        @close="sidebarOpen = false"
        @retry="retryDevices"
      />
      <FleetMapPanel
        :devices="devices"
        :selected-id="selectedId"
        :sidebar-open="sidebarOpen"
        :loading="isPending"
        :error="!!error"
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
