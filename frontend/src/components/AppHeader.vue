<script setup lang="ts">
import { MapPin } from "@lucide/vue";
</script>

<template>
  <header class="topbar">
    <div class="brand-block">
      <div class="brand-symbol"><MapPin :size="19" :stroke-width="2.6" /></div>
      <div class="brand-name">One Step <span>GPS</span></div>
    </div>
    <div class="topbar-right">
      <span class="demo-label"><i /> DEMO DATA</span>
      <div class="avatar">OS</div>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 72px;
  padding: 0 28px;
  background: #12253c;
  color: #f6f8fb;
  box-shadow: 0 2px 15px #0d1c2b25;
}
.brand-block,
.topbar-right {
  display: flex;
  align-items: center;
}
.brand-symbol {
  width: 37px;
  height: 37px;
  display: grid;
  place-items: center;
  border-radius: 11px;
  background: #e96743;
  color: #fff;
  margin-right: 11px;
  box-shadow: 0 4px 12px #09111d55;
}
.brand-name {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.04em;
  white-space: nowrap;
}
.brand-name span {
  color: #ff9a75;
}

.topbar-right {
  gap: 20px;
}
.demo-label {
  color: #cbd6e1;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.12em;
  display: flex;
  align-items: center;
  gap: 8px;
}
.demo-label i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f1ba63;
  box-shadow: 0 0 0 4px #f1ba6324;
}
.avatar {
  display: grid;
  place-items: center;
  width: 35px;
  height: 35px;
  border: 1px solid #7b8c9e;
  border-radius: 50%;
  background: #42566d;
  color: #fff;
  font-size: 11px;
  font-weight: 800;
}
@media (max-width: 760px) {
  .topbar {
    height: 62px;
    padding: 0 16px;
  }
  .demo-label {
    display: none;
  }
}
</style>
