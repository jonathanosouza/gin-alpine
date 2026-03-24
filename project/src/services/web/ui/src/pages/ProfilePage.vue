<template>
  <AppShell>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Perfil</h1>
        <p class="text-sm text-muted-foreground">Atualize suas informações.</p>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="rounded-2xl border border-border bg-card p-5">
          <div class="text-sm font-semibold">Dados</div>
          <div class="mt-4 space-y-4">
            <div class="space-y-2">
              <label class="text-sm font-medium">Nome</label>
              <input
                v-model.trim="form.name"
                type="text"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium">Email</label>
              <input
                v-model.trim="form.email"
                type="email"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium">Nova senha (opcional)</label>
              <input
                v-model="form.password"
                type="password"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>

          <div v-if="error" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {{ error }}
          </div>
          <div v-if="success" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
            {{ success }}
          </div>

          <div class="mt-5 flex justify-end">
            <button
              type="button"
              class="rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              :disabled="saving"
              @click="save"
            >
              Salvar
            </button>
          </div>
        </div>

        <div class="rounded-2xl border border-border bg-card p-5">
          <div class="text-sm font-semibold">Resumo</div>
          <div class="mt-4 space-y-3 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-muted-foreground">ID</span>
              <span class="font-medium">{{ ctx.User?.ID ?? "-" }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-muted-foreground">Nome</span>
              <span class="font-medium">{{ ctx.User?.Name ?? "-" }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-muted-foreground">Email</span>
              <span class="font-medium">{{ ctx.User?.Email ?? "-" }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import AppShell from "../components/AppShell.vue";
import { getAppContext } from "../lib/appContext";
import { apiFetch } from "../lib/api";

const ctx = getAppContext();
const saving = ref(false);
const error = ref("");
const success = ref("");

const form = reactive({
  name: String(ctx.User?.Name ?? ""),
  email: String(ctx.User?.Email ?? ""),
  password: "",
});

async function save() {
  if (saving.value) return;
  error.value = "";
  success.value = "";
  const id = ctx.User?.ID;
  if (!id) {
    error.value = "Usuário não encontrado na sessão.";
    return;
  }
  saving.value = true;
  const payload: Record<string, unknown> = {
    name: form.name,
    email: form.email,
  };
  if (form.password) {
    payload.password = form.password;
  }
  const res = await apiFetch(`/api/users/${id}`, { method: "PUT", json: payload });
  if (!res.ok) {
    if (res.raw && typeof res.raw === "object" && "errors" in (res.raw as any)) {
      const errors = (res.raw as any).errors;
      const first = errors && typeof errors === "object" ? Object.values(errors)[0] : undefined;
      if (typeof first === "string") {
        error.value = first;
      } else {
        error.value = res.error;
      }
    } else {
      error.value = res.error;
    }
    saving.value = false;
    return;
  }
  success.value = "Perfil atualizado.";
  form.password = "";
  saving.value = false;
}
</script>
