<script setup lang="ts">
import type { Component } from "vue";
import { Card } from "@/components/ui/card";

defineProps<{
  value: number | string;
  label: string;
  icon: Component;
  tone?: "blue" | "green" | "orange";
}>();
</script>

<template>
  <Card class="summary-card">
    <span class="summary-icon" :class="tone ?? 'blue'"
      ><component :is="icon" :size="16"
    /></span>
    <strong>{{ value }}</strong>
    <small>{{ label }}</small>
  </Card>
</template>

<style scoped>
.summary-card {
  display: flex;
  flex-direction: column;
  gap: 0;
  min-width: 0;
  min-height: 101px;
  padding: 13px;
  border: 1px solid #e6ebf0;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 3px 12px #1f354e08;
}
.summary-icon {
  display: grid;
  place-items: center;
  width: 25px;
  height: 25px;
  border-radius: 7px;
  margin-bottom: 6px;
}
.summary-icon.blue {
  background: #e9eef6;
  color: #315c87;
}
.summary-icon.green {
  background: #e5f5e9;
  color: #28905d;
}
.summary-icon.orange {
  background: #fff0e8;
  color: #dc7448;
}
strong {
  font-size: 20px;
  line-height: 1;
  font-weight: 800;
  letter-spacing: -0.05em;
}
small {
  margin-top: 4px;
  font-size: 10px;
  color: #8a98a8;
  white-space: nowrap;
}
@media (max-width: 1060px) {
  .summary-card {
    padding: 10px;
  }
  small {
    font-size: 9px;
  }
}
</style>
