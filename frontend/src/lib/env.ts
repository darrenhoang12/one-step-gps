import { z } from "zod";

const optionalNonEmptyString = z.preprocess(
  (value) => (typeof value === "string" && value.trim() === "" ? undefined : value),
  z.string().trim().min(1).optional(),
);

const envSchema = z.object({
  VITE_GOOGLE_MAPS_API_KEY: optionalNonEmptyString,
  VITE_GOOGLE_MAPS_MAP_ID: z.preprocess(
    (value) => (typeof value === "string" && value.trim() === "" ? undefined : value),
    z.string().trim().min(1).default("DEMO_MAP_ID"),
  ),
});

const parsed = envSchema.parse(import.meta.env);

export const env = {
  googleMapsApiKey: parsed.VITE_GOOGLE_MAPS_API_KEY,
  googleMapsMapId: parsed.VITE_GOOGLE_MAPS_MAP_ID,
};
