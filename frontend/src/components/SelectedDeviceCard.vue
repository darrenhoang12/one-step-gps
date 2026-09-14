<script setup lang="ts">
import { Gauge, MapPin, Truck, Wifi } from "@lucide/vue";
import { Card } from "@/components/ui/card";
import { formatCoordinates } from "@/lib/deviceFormat";
import type { Device } from "@/types/device";

defineProps<{ device: Device }>();
</script>

<template>
  <Card class="selected-card">
    <span class="selected-icon"><Truck :size="23" /></span>
    <div class="selected-main">
      <small>SELECTED DEVICE</small><strong>{{ device.display_name }}</strong
      ><span><MapPin :size="13" /> {{ formatCoordinates(device) }}</span>
    </div>
    <div class="selected-stats">
      <div>
        <Gauge :size="15" /><strong>{{ Math.round(device.speed) }} mph</strong
        ><small>Speed</small>
      </div>
      <div>
        <Wifi :size="15" /><strong>{{
          device.online ? "Online" : "Offline"
        }}</strong
        ><small>Status</small>
      </div>
    </div>
  </Card>
</template>

<style scoped>
.selected-card {
  display: flex;
  flex-direction: row;
  gap: 0;
  align-items: center;
  min-width: 0;
  width: min(100%, 466px);
  padding: 16px;
  border: 1px solid #ffffffd0;
  border-radius: 15px;
  background: #fffffff2;
  box-shadow: 0 12px 35px #1a35502a;
  backdrop-filter: blur(12px);
}
.selected-icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 46px;
  height: 46px;
  border-radius: 11px;
  background: #eaf1f8;
  color: #2b5d85;
}
.selected-main {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  padding-left: 13px;
}
.selected-main small {
  font-size: 8px;
  font-weight: 800;
  letter-spacing: 0.16em;
  color: #e06b49;
}
.selected-main strong {
  margin-top: 3px;
  font-size: 15px;
  font-weight: 800;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.selected-main span {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  font-size: 10px;
  color: #91a0ad;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.selected-stats {
  display: flex;
  gap: 16px;
  border-left: 1px solid #e6ebf1;
  padding-left: 15px;
}
.selected-stats > div {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 3px;
  color: #66809a;
}
.selected-stats strong {
  color: #2d435b;
  font-size: 11px;
  white-space: nowrap;
}
.selected-stats small {
  color: #9ca9b5;
  font-size: 9px;
}
@media (max-width: 760px) {
  .selected-card {
    width: 100%;
  }
  .selected-stats {
    gap: 9px;
    padding-left: 10px;
  }
  .selected-icon {
    width: 38px;
    height: 38px;
  }
}
@media (max-width: 420px) {
  .selected-stats > div:last-child {
    display: none;
  }
  .selected-card {
    padding: 12px;
  }
}
</style>
