<template>
  <AppShell>
    <div class="space-y-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Usuários</h1>
          <p class="text-sm text-muted-foreground">Gerencie acessos ao sistema.</p>
        </div>

        <button
          type="button"
          class="inline-flex items-center justify-center rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
          @click="openCreate = true"
        >
          Novo usuário
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <input
          v-model.trim="term"
          type="search"
          placeholder="Buscar (mín. 3 caracteres)"
          class="w-full max-w-md rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
          @keyup.enter="load()"
        />
        <button
          type="button"
          class="rounded-xl border border-border bg-card px-4 py-2 text-sm font-medium hover:bg-secondary"
          @click="load()"
        >
          Buscar
        </button>
      </div>

      <div class="rounded-2xl border border-border bg-card">
        <div class="flex items-center justify-between border-b border-border px-4 py-3">
          <div class="text-sm font-semibold">Lista</div>
          <div class="text-xs text-muted-foreground">
            Página {{ page }} de {{ totalPages || 1 }}
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase tracking-wide text-muted-foreground">
              <tr class="border-b border-border">
                <th class="px-4 py-3">Nome</th>
                <th class="px-4 py-3">Email</th>
                <th class="px-4 py-3">Cargo</th>
                <th class="px-4 py-3">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td class="px-4 py-4 text-muted-foreground" colspan="4">Carregando...</td>
              </tr>
              <tr v-else-if="items.length === 0">
                <td class="px-4 py-4 text-muted-foreground" colspan="4">Sem dados.</td>
              </tr>
              <tr v-for="u in items" :key="u.id" class="border-b border-border last:border-0">
                <td class="px-4 py-3 font-medium">{{ u.name }}</td>
                <td class="px-4 py-3">{{ u.email }}</td>
                <td class="px-4 py-3">{{ u.role || "-" }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold"
                    :class="u.enabled ? 'bg-emerald-100 text-emerald-800' : 'bg-red-100 text-red-800'"
                  >
                    {{ u.enabled ? "Ativo" : "Desabilitado" }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-2 border-t border-border px-4 py-3">
          <div class="text-xs text-muted-foreground">Total: {{ totalItems }}</div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="rounded-lg border border-border px-3 py-1.5 text-sm font-medium hover:bg-secondary disabled:opacity-50"
              :disabled="page <= 1 || loading"
              @click="prev()"
            >
              Anterior
            </button>
            <button
              type="button"
              class="rounded-lg border border-border px-3 py-1.5 text-sm font-medium hover:bg-secondary disabled:opacity-50"
              :disabled="page >= totalPages || loading"
              @click="next()"
            >
              Próxima
            </button>
          </div>
        </div>
      </div>

      <div v-show="openCreate" class="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div class="absolute inset-0 bg-black/40" @click="closeCreate" />
        <div class="relative w-full max-w-lg rounded-2xl border border-border bg-card p-6 shadow-lg">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-lg font-semibold">Novo usuário</div>
              <div class="text-sm text-muted-foreground">Cria um acesso para o BI.</div>
            </div>
            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border hover:bg-secondary"
              @click="closeCreate"
            >
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18" />
                <path d="M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div v-if="createError" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {{ createError }}
          </div>

          <div class="mt-5 grid gap-4 sm:grid-cols-2">
            <div class="space-y-2 sm:col-span-2">
              <label class="text-sm font-medium">Nome</label>
              <input
                v-model.trim="create.name"
                type="text"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div class="space-y-2 sm:col-span-2">
              <label class="text-sm font-medium">Email</label>
              <input
                v-model.trim="create.email"
                type="email"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium">Cargo</label>
              <select
                v-model.number="create.role_id"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              >
                <option :value="1">Customer</option>
                <option :value="2">Manager</option>
                <option :value="3">Admin</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium">Senha</label>
              <input
                v-model="create.password"
                type="text"
                class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>

          <div class="mt-6 flex items-center justify-end gap-2">
            <button
              type="button"
              class="rounded-xl border border-border bg-card px-4 py-2 text-sm font-medium hover:bg-secondary"
              :disabled="creating"
              @click="closeCreate"
            >
              Cancelar
            </button>
            <button
              type="button"
              class="rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              :disabled="creating"
              @click="submitCreate"
            >
              Criar
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import AppShell from "../components/AppShell.vue";
import { apiFetch } from "../lib/api";

type UserItem = {
  id: number;
  name: string;
  email: string;
  role: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  total_count: number;
};

type UsersResponse = {
  page: number;
  items_per_page: number;
  total_items: number;
  total_pages: number;
  items: UserItem[];
};

const loading = ref(false);
const creating = ref(false);
const term = ref("");
const page = ref(1);
const limit = ref(10);
const totalPages = ref(0);
const totalItems = ref(0);
const items = ref<UserItem[]>([]);

const openCreate = ref(false);
const createError = ref("");
const create = reactive({
  name: "",
  email: "",
  role_id: 1,
  password: "",
});

function closeCreate() {
  openCreate.value = false;
  createError.value = "";
}

async function load() {
  loading.value = true;
  const qp = new URLSearchParams({
    page: String(page.value),
    limit: String(limit.value),
  });
  let url = `/api/users?${qp.toString()}`;
  const t = term.value.trim();
  if (t.length >= 3) {
    const sp = new URLSearchParams({
      page: String(page.value),
      limit: String(limit.value),
      term: t,
    });
    url = `/api/users/search?${sp.toString()}`;
  }

  const res = await apiFetch<UsersResponse>(url, { method: "GET" });
  if (res.ok) {
    items.value = res.data.items ?? [];
    totalPages.value = res.data.total_pages ?? 0;
    totalItems.value = res.data.total_items ?? 0;
  } else {
    items.value = [];
    totalPages.value = 0;
    totalItems.value = 0;
  }
  loading.value = false;
}

function prev() {
  if (page.value <= 1) return;
  page.value -= 1;
  void load();
}

function next() {
  if (page.value >= totalPages.value) return;
  page.value += 1;
  void load();
}

async function submitCreate() {
  if (creating.value) return;
  createError.value = "";
  creating.value = true;
  const payload = {
    name: create.name,
    email: create.email,
    password: create.password,
    role_id: create.role_id,
  };
  const res = await apiFetch<unknown>("/api/users/", { method: "POST", json: payload });
  if (!res.ok) {
    if (res.raw && typeof res.raw === "object" && "errors" in (res.raw as any)) {
      const errors = (res.raw as any).errors;
      const first = errors && typeof errors === "object" ? Object.values(errors)[0] : undefined;
      if (typeof first === "string") {
        createError.value = first;
      } else {
        createError.value = res.error;
      }
    } else {
      createError.value = res.error;
    }
    creating.value = false;
    return;
  }

  closeCreate();
  create.name = "";
  create.email = "";
  create.password = "";
  create.role_id = 1;
  page.value = 1;
  await load();
  creating.value = false;
}

onMounted(() => {
  void load();
});
</script>
