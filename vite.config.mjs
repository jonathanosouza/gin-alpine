import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";

export default defineConfig({
  plugins: [vue()],
  root: path.resolve("project/src/services/web/ui"),
  base: "/static/vue/",
  server: {
    fs: {
      allow: [
        path.resolve("."),
        path.resolve("project/src/services/web/static"),
      ],
    },
  },
  build: {
    outDir: path.resolve("project/src/services/web/static/vue"),
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      output: {
        entryFileNames: "app.js",
        chunkFileNames: "chunk-[name].js",
        assetFileNames: "[name][extname]",
      },
    },
  },
});
