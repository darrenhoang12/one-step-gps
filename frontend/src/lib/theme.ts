import { ref } from "vue";

export type Theme = "light" | "dark";

const storageKey = "one-step-gps-theme";
const theme = ref<Theme>("light");

function preferredTheme(): Theme {
  const saved = localStorage.getItem(storageKey);
  if (saved === "light" || saved === "dark") return saved;
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function applyTheme(value: Theme) {
  theme.value = value;
  document.documentElement.classList.toggle("dark", value === "dark");
  document.documentElement.style.colorScheme = value;
}

export function initializeTheme() {
  applyTheme(preferredTheme());
}

export function useTheme() {
  function toggleTheme() {
    const next = theme.value === "dark" ? "light" : "dark";
    localStorage.setItem(storageKey, next);
    applyTheme(next);
  }

  return { theme, toggleTheme };
}
