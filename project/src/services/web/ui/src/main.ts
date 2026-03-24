import { createApp } from "vue";
import App from "./app/App.vue";
import { router } from "./router";
import "../../static/css/styles.css";

async function bootstrap() {
  const w = window as unknown as { __APP_CONTEXT__?: unknown };
  try {
    const apiBase = window.location.port === "3000" ? "" : "http://localhost:3000";
    const res = await fetch(`${apiBase}/api/context`, {
      headers: {
        Accept: "application/json",
      },
      credentials: "include",
    });
    if (res.ok) {
      w.__APP_CONTEXT__ = await res.json().catch(() => ({}));
    }
  } catch {
    w.__APP_CONTEXT__ = w.__APP_CONTEXT__ ?? {};
  }

  createApp(App).use(router).mount("#app");
}

void bootstrap();
