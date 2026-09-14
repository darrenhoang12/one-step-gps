<script setup lang="ts">
import { Search } from "@lucide/vue";
import { Input } from "@/components/ui/input";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

export type DeviceFilter = "all" | "online" | "offline";

defineProps<{ search: string; filter: DeviceFilter }>();
const emit = defineEmits<{
  "update:search": [value: string];
  "update:filter": [value: DeviceFilter];
}>();
const options: { value: DeviceFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "online", label: "Online" },
  { value: "offline", label: "Offline" },
];
</script>

<template>
  <label class="search-box">
    <Search :size="17" />
    <Input
      :model-value="search"
      type="search"
      placeholder="Search devices"
      aria-label="Search devices"
      @update:model-value="emit('update:search', String($event))"
    />
  </label>
  <ToggleGroup
    class="filter-row"
    type="single"
    :model-value="filter"
    aria-label="Filter devices"
    @update:model-value="
      emit('update:filter', ($event || 'all') as DeviceFilter)
    "
  >
    <ToggleGroupItem
      v-for="option in options"
      :key="option.value"
      :value="option.value"
      :class="{ active: filter === option.value }"
    >
      {{ option.label }}
    </ToggleGroupItem>
  </ToggleGroup>
</template>

<style scoped>
.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 17px;
  padding: 0 12px;
  height: 41px;
  border: 1px solid #dbe3eb;
  border-radius: 9px;
  background: #fff;
  color: #91a0ae;
}
.search-box:focus-within {
  border-color: #6b9ac0;
  box-shadow: 0 0 0 3px #dceaf6;
}
.search-box input {
  min-width: 0;
  flex: 1;
  outline: none;
  border: 0;
  background: transparent;
  color: #213248;
  font-size: 12px;
}
.search-box input::placeholder {
  color: #9aa6b2;
}
.filter-row {
  display: inline-flex;
  gap: 4px;
  margin: 15px 0;
  padding: 4px;
  border: 1px solid #dbe3eb;
  border-radius: 10px;
  background: #eaf0f5;
}
.filter-row [data-slot="toggle-group-item"] {
  border: 1px solid transparent;
  border-radius: 7px;
  min-height: 34px;
  padding: 7px 14px;
  color: #51667c;
  background: transparent;
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;
}
.filter-row [data-slot="toggle-group-item"].active {
  background: #223e5f;
  color: #fff;
  box-shadow: 0 2px 6px #112f5230;
}
.filter-row [data-slot="toggle-group-item"]:not(.active):hover {
  background: #fff;
  color: #223e5f;
}
</style>
