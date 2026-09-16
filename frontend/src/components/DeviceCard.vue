<script setup lang="ts">
import { Archive, ArchiveRestore, Clock3, Eye, EyeOff, Gauge, GripVertical, ImageOff, MapPin, Pencil, RotateCcw, Truck, Upload } from "@lucide/vue";
import { ref } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import ImageCropDialog from "@/components/ImageCropDialog.vue";
import { formatCoordinates, formatReportedAt } from "@/lib/deviceFormat";
import type { Device } from "@/types/device";

const props = defineProps<{ device: Device; selected?: boolean; draggable?: boolean; saving?: boolean }>();
const emit = defineEmits<{
  select: [deviceId: string];
  update: [deviceId: string, changes: Partial<Device>];
  upload: [deviceId: string, file: File];
  removeIcon: [deviceId: string];
  dragStart: [deviceId: string];
  drop: [deviceId: string];
}>();
const editing = ref(false);
const customName = ref("");
const fileInput = ref<HTMLInputElement | null>(null);
const imageToCrop = ref<File | null>(null);
const fileError = ref<string | null>(null);

function beginEditing() {
  customName.value = props.device.custom_display_name ?? props.device.display_name;
  editing.value = true;
}

function saveName() {
  emit("update", props.device.device_id, { custom_display_name: customName.value.trim() || null });
  editing.value = false;
}

function chooseFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  fileError.value = null;
  if (file && file.size > 5 * 1024 * 1024) {
    fileError.value = "Icon must be 5 MB or smaller";
  } else if (file) {
    imageToCrop.value = file;
  }
  (event.target as HTMLInputElement).value = "";
}

function uploadCrop(file: File) {
  imageToCrop.value = null;
  emit("upload", props.device.device_id, file);
}
</script>

<template>
  <Card
    class="device-card"
    :class="{ selected }"
    role="button"
    tabindex="0"
    :aria-pressed="selected"
    :draggable="draggable"
    @click="emit('select', device.device_id)"
    @keydown.enter="emit('select', device.device_id)"
    @keydown.space.prevent="emit('select', device.device_id)"
    @dragstart="emit('dragStart', device.device_id)"
    @dragover.prevent
    @drop.prevent="emit('drop', device.device_id)"
  >
    <span class="card-top">
      <GripVertical v-if="draggable" class="drag-handle" :size="16" />
      <span class="device-icon">
        <img v-if="device.icon_url" :src="device.icon_url" alt="" />
        <Truck v-else :size="19" />
      </span>
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
    <form v-if="editing" class="name-editor" @click.stop @submit.prevent="saveName">
      <Input v-model="customName" maxlength="80" aria-label="Custom device name" autofocus />
      <Button type="submit" size="sm">Save</Button>
      <Button type="button" size="sm" variant="ghost" @click="editing = false">Cancel</Button>
    </form>
    <span v-else class="card-actions" @click.stop>
      <Button variant="ghost" size="icon-sm" title="Rename device" @click="beginEditing"><Pencil :size="14" /></Button>
      <Button
        v-if="device.custom_display_name"
        variant="ghost"
        size="icon-sm"
        title="Reset to original name"
        aria-label="Reset to original device name"
        :disabled="saving"
        @click="emit('update', device.device_id, { custom_display_name: null })"
      >
        <RotateCcw :size="14" />
      </Button>
      <Button variant="ghost" size="icon-sm" :title="device.hidden ? 'Show device' : 'Hide device'" @click="emit('update', device.device_id, { hidden: !device.hidden })">
        <Eye v-if="device.hidden" :size="14" /><EyeOff v-else :size="14" />
      </Button>
      <Button variant="ghost" size="icon-sm" :title="device.archived ? 'Restore device' : 'Archive device'" @click="emit('update', device.device_id, { archived: !device.archived })">
        <ArchiveRestore v-if="device.archived" :size="14" /><Archive v-else :size="14" />
      </Button>
      <Button variant="ghost" size="icon-sm" title="Upload device icon (max 5 MB, 1024×1024)" @click="fileInput?.click()"><Upload :size="14" /></Button>
      <Button
        v-if="device.icon_url"
        variant="ghost"
        size="icon-sm"
        title="Remove device icon"
        aria-label="Remove device icon"
        :disabled="saving"
        @click="emit('removeIcon', device.device_id)"
      ><ImageOff :size="14" /></Button>
      <input ref="fileInput" class="file-input" type="file" accept="image/png,image/jpeg,image/webp,image/gif" @change="chooseFile" />
      <span v-if="saving" class="saving-label">Saving…</span>
    </span>
    <span v-if="fileError" class="file-error" role="alert">{{ fileError }}</span>
    <ImageCropDialog :file="imageToCrop" @close="imageToCrop = null" @crop="uploadCrop" />
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
.device-card[draggable="true"] {
  cursor: grab;
}
.device-card[draggable="true"]:active {
  cursor: grabbing;
}
.drag-handle {
  flex: none;
  color: #9aa8b6;
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
.device-icon img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
}
.file-error {
  display: block;
  margin-top: 8px;
  color: #b33b30;
  font-size: 10px;
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
.card-actions,
.name-editor {
  display: flex;
  align-items: center;
  gap: 3px;
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid #edf1f4;
}
.name-editor input {
  height: 32px;
  font-size: 11px;
}
.file-input {
  display: none;
}
.saving-label {
  margin-left: auto;
  color: #7c8b9d;
  font-size: 9px;
  font-weight: 700;
}
.card-location {
  margin-top: 10px;
  color: #a0aab4;
  font-size: 10px;
}
</style>
