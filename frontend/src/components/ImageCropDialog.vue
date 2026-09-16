<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { Cropper } from "vue-advanced-cropper";
import "vue-advanced-cropper/dist/style.css";
import { Button } from "@/components/ui/button";

type CropperResult = { canvas?: HTMLCanvasElement | null };
type CropperInstance = { getResult: () => CropperResult };

const props = defineProps<{ file: File | null }>();
const emit = defineEmits<{
  close: [];
  crop: [file: File];
}>();

const cropper = ref<CropperInstance | null>(null);
const saving = ref(false);
const error = ref<string | null>(null);
const sourceUrl = ref("");
const fileName = computed(() => {
  const base = props.file?.name.replace(/\.[^.]+$/, "") || "device-icon";
  return `${base}.webp`;
});

function releaseSource() {
  if (sourceUrl.value) URL.revokeObjectURL(sourceUrl.value);
  sourceUrl.value = "";
}

watch(
  () => props.file,
  (file) => {
    releaseSource();
    error.value = null;
    if (file) sourceUrl.value = URL.createObjectURL(file);
  },
  { immediate: true },
);

onBeforeUnmount(releaseSource);

async function saveCrop() {
  const source = cropper.value?.getResult().canvas;
  if (!source) {
    error.value = "Could not crop this image";
    return;
  }

  saving.value = true;
  error.value = null;
  try {
    const output = document.createElement("canvas");
    output.width = 512;
    output.height = 512;
    const context = output.getContext("2d");
    if (!context) throw new Error("Could not prepare the cropped icon");
    context.drawImage(source, 0, 0, 512, 512);
    const blob = await new Promise<Blob>((resolve, reject) =>
      output.toBlob(
        (value) =>
          value ? resolve(value) : reject(new Error("Could not create the cropped icon")),
        "image/webp",
        0.9,
      ),
    );
    emit("crop", new File([blob], fileName.value, { type: "image/webp" }));
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "Could not crop this image";
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="file" class="crop-backdrop" role="presentation" @click.self="emit('close')">
      <section class="crop-dialog" role="dialog" aria-modal="true" aria-labelledby="crop-title">
        <header>
          <div>
            <h2 id="crop-title">Crop device icon</h2>
            <p>Drag and zoom the image inside the square.</p>
          </div>
          <button class="close-button" type="button" aria-label="Close crop dialog" @click="emit('close')">×</button>
        </header>

        <Cropper
          v-if="sourceUrl"
          ref="cropper"
          class="cropper"
          :src="sourceUrl"
          :stencil-props="{ aspectRatio: 1 }"
          :resize-image="{ adjustStencil: false }"
          image-restriction="stencil"
        />

        <p v-if="error" class="crop-error" role="alert">{{ error }}</p>
        <footer>
          <span>Exports as 512×512 WebP</span>
          <div>
            <Button type="button" variant="ghost" :disabled="saving" @click="emit('close')">Cancel</Button>
            <Button type="button" :disabled="saving" @click="saveCrop">{{ saving ? "Preparing…" : "Use icon" }}</Button>
          </div>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.crop-backdrop {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 20px;
  background: rgb(13 29 45 / 62%);
  backdrop-filter: blur(4px);
}
.crop-dialog {
  width: min(92vw, 560px);
  overflow: hidden;
  border: 1px solid #dce4eb;
  border-radius: 16px;
  background: white;
  box-shadow: 0 24px 70px rgb(11 29 48 / 30%);
}
header,
footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px;
}
h2 {
  margin: 0;
  color: #213951;
  font-size: 18px;
}
p {
  margin: 4px 0 0;
  color: #718194;
  font-size: 12px;
}
.close-button {
  border: 0;
  background: transparent;
  color: #718194;
  cursor: pointer;
  font-size: 26px;
}
.cropper {
  height: min(58vh, 430px);
  background: #111b25;
}
footer {
  border-top: 1px solid #e7edf2;
}
footer > span {
  color: #718194;
  font-size: 11px;
}
footer > div {
  display: flex;
  gap: 8px;
}
.crop-error {
  margin: 12px 20px 0;
  color: #b33b30;
}
@media (max-width: 520px) {
  footer {
    align-items: stretch;
    flex-direction: column;
  }
  footer > div {
    justify-content: flex-end;
  }
}
</style>
