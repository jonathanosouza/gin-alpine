import { getCSRFToken } from "./appContext";

export async function apiFetch<T>(
  url: string,
  init?: RequestInit & { json?: unknown },
): Promise<{ ok: true; data: T } | { ok: false; status: number; error: string; raw?: unknown }> {
  const headers = new Headers(init?.headers ?? {});
  headers.set("Accept", "application/json");
  if (init?.json !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  const csrf = getCSRFToken();
  if (csrf) {
    headers.set("X-CSRF-Token", csrf);
  }

  let res: Response;
  try {
    const base = window.location.port === "3000" ? "" : "http://localhost:3000";
    res = await fetch(base + url, {
      ...init,
      headers,
      body: init?.json !== undefined ? JSON.stringify(init.json) : init?.body,
      credentials: "include",
    });
  } catch (e) {
    return { ok: false, status: 0, error: "Falha de conexão com o servidor.", raw: e };
  }

  const contentType = res.headers.get("content-type") ?? "";
  const isJSON = contentType.includes("application/json");
  const raw = isJSON ? await res.json().catch(() => undefined) : await res.text().catch(() => undefined);

  if (res.ok) {
    return { ok: true, data: raw as T };
  }

  const msg =
    typeof raw === "object" && raw && "error" in raw && typeof (raw as any).error === "string"
      ? String((raw as any).error)
      : "Erro ao processar sua solicitação.";

  return { ok: false, status: res.status, error: msg, raw };
}
