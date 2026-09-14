<script setup lang="ts">
import { Clock3, Gauge, MapPin, Truck } from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { formatCoordinates, formatReportedAt } from "@/lib/deviceFormat";
import type { Device } from "@/types/device";

defineProps<{ device: Device; selected?: boolean }>();
const emit = defineEmits<{ select: [deviceId: string] }>();
</script>

<template>
  <Card
    class="device-card"
    :class="{ selected }"
    role="button"
    tabindex="0"
    :aria-pressed="selected"
    @click="emit('select', device.device_id)"
    @keydown.enter="emit('select', device.device_id)"
    @keydown.space.prevent="emit('select', device.device_id)"
  >
    <span class="card-top">
      <span class="device-icon"><Truck :size="19" /></span>
      <span class="device-title"
        ><strong>{{ device.display_name }}</strong
        ><small>{{ device.model }}</small></span
      >
      <Badge class="status-badge" :class="device.online ? 'online' : 'offline'"
        ><i />{{ device.online ? "Online" : "Offline" }}</Badge
      >
    </span>
    <span class="card-divider" />
    <span class="card-data">
      <span><Gauge :size="14" /> {{ Math.round(device.speed) }} mph</span>
      <span
        ><Clock3 :size="14" />
        {{ formatReportedAt(device.last_reported_at) }}</span
      >
    </span>
    <span class="card-location"
      ><MapPin :size="13" /> {{ formatCoordinates(device) }}</span
    >
  </Card>
</template>

<style scoped>
.device-card {
  width: 100%;
  gap: 0;
  padding: 15px;
  text-align: left;
  border: 1px solid #e3e9ef;
  border-radius: 13px;
  background: #fff;
  box-shadow: 0 3px 10px #1b355008;
  cursor: pointer;
  color: inherit;
  transition:
    border-color 0.18s,
    box-shadow 0.18s,
    transform 0.18s;
}
.device-card:hover {
  border-color: #9cb5cf;
  box-shadow: 0 5px 17px #1b355018;
  transform: translateY(-1px);
}
.device-card.selected {
  border-color: #41749f;
  box-shadow:
    inset 3px 0 #e66843,
    0 5px 16px #2149711a;
}
.device-card:focus-visible {
  outline: 2px solid #41749f;
  outline-offset: 2px;
}
.card-top {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 10px;
}
.device-icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: #e8eff6;
  color: #3d6488;
}
.device-title {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  gap: 3px;
}
.device-title strong {
  font-size: 12px;
  font-weight: 800;
  color: #233951;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.device-title small {
  font-size: 10px;
  color: #8a99a8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.status-badge {
  display: flex;
  align-items: center;
  gap: 5px;
  flex: none;
  padding: 5px 7px;
  border-radius: 6px;
  font-size: 9px;
  font-weight: 800;
}
.status-badge i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.status-badge.online {
  color: #2a9660;
  background: #e7f6ee;
}
.status-badge.offline {
  color: #8a97a2;
  background: #edf1f4;
}
.card-divider {
  display: block;
  height: 1px;
  margin: 13px 0 11px;
  background: #edf1f4;
}
.card-data {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  color: #6b7e91;
  font-size: 10px;
  font-weight: 600;
}
.card-data span,
.card-location {
  display: flex;
  align-items: center;
  gap: 5px;
}
.card-location {
  margin-top: 10px;
  color: #a0aab4;
  font-size: 10px;
}
</style>
