import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "localhost",
    port: 5174,
    strictPort: true,
    hmr: {
      host: "localhost",
      port: 5174,
    },
  },
  preview: {
    host: "localhost",
    port: 4174,
    strictPort: true,
  },
});
