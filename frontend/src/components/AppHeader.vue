<script setup lang="ts">
import { Moon, Sun } from "@lucide/vue";
import { useTheme } from "@/lib/theme";

const { theme, toggleTheme } = useTheme();
</script>

<template>
  <header class="topbar">
    <div class="brand-block">
      <img
        class="brand-logo"
        src="@/assets/osg-logo.png"
        alt="OneStepGPS"
      />
    </div>
    <div class="topbar-right">
      <button
        class="theme-toggle"
        type="button"
        :aria-label="`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`"
        :title="`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`"
        :aria-pressed="theme === 'dark'"
        @click="toggleTheme"
      >
        <Sun v-if="theme === 'dark'" :size="18" />
        <Moon v-else :size="18" />
      </button>
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
.brand-logo {
  display: block;
  width: 190px;
  height: auto;
  padding: 7px 10px;
  border-radius: 10px;
  background: #fff;
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
.theme-toggle {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border: 1px solid #526a82;
  border-radius: 10px;
  background: #243b54;
  color: #f7c96f;
  cursor: pointer;
  transition: background 0.18s, border-color 0.18s, transform 0.18s;
}
.theme-toggle:hover {
  border-color: #7e98b2;
  background: #304a65;
  transform: translateY(-1px);
}
.theme-toggle:focus-visible {
  outline: 2px solid #76b8ee;
  outline-offset: 2px;
}
@media (max-width: 760px) {
  .topbar {
    height: 62px;
    padding: 0 16px;
  }
  .brand-logo {
    width: 158px;
    padding: 6px 9px;
  }
  .demo-label {
    display: none;
  }
}
</style>
