import { z } from "zod";

const envSchema = z.object({
  VITE_GOOGLE_MAPS_API_KEY: z.string().trim().min(1, "VITE_GOOGLE_MAPS_API_KEY is required"),
  VITE_GOOGLE_MAPS_MAP_ID: z.preprocess(
    (value) => (typeof value === "string" && value.trim() === "" ? undefined : value),
    z.string().trim().min(1).default("DEMO_MAP_ID"),
  ),
  VITE_API_BASE_URL: z.preprocess(
    (value) => (typeof value === "string" && value.trim() === "" ? undefined : value),
    z.url().refine((value) => {
      const url = new URL(value);
      return (url.protocol === "http:" || url.protocol === "https:") && !url.search && !url.hash;
    }, "VITE_API_BASE_URL must be an HTTP(S) URL without a query or fragment").optional(),
  ),
});

const parsed = envSchema.parse(import.meta.env);

export const env = {
  googleMapsApiKey: parsed.VITE_GOOGLE_MAPS_API_KEY,
  googleMapsMapId: parsed.VITE_GOOGLE_MAPS_MAP_ID,
  apiBaseUrl: parsed.VITE_API_BASE_URL?.replace(/\/+$/, "") ?? "",
};
