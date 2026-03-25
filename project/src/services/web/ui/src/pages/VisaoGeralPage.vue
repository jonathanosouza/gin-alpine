<template>
  <AppShell>
    <div class="space-y-6">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Visão Geral</h1>
          <p class="text-sm text-muted-foreground">
            Uma visão consolidada para reduzir redundâncias e maximizar performance.
          </p>
        </div>
        <RouterLink
          to="/visao-geral/calculos"
          class="inline-flex items-center gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm hover:bg-blue-50"
        >
          Documentação de cálculos
        </RouterLink>
      </div>

      <DashboardFilters
        v-model="filters"
        :filiais-options="filiais"
        :loading="loading"
        @change="onFiltersChanged"
        @submit="reload"
      />

      <div v-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {{ error }}
        </div>
      <div v-else-if="warning" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
        {{ warning }}
      </div>

      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Faturamento total</div>
          <div class="mt-2 text-2xl font-bold">{{ formatMoney(dashboard?.metrics.faturamentoTotal ?? 0) }}</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Clientes ativos</div>
          <div class="mt-2 text-2xl font-bold">{{ formatInt(dashboard?.metrics.clientesAtivos ?? 0) }}</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Notas fiscais</div>
          <div class="mt-2 text-2xl font-bold">{{ formatInt(dashboard?.metrics.totalNFs ?? 0) }}</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Ticket médio</div>
          <div class="mt-2 text-2xl font-bold">{{ formatMoney(dashboard?.metrics.ticketMedio ?? 0) }}</div>
        </div>
      </div>

      <div class="rounded-2xl border border-border bg-card p-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm font-semibold">Vendas mensais</div>
            <div class="text-xs text-muted-foreground">Evolução mês a mês no período filtrado</div>
          </div>
          <div class="text-xs text-muted-foreground">
            <span v-if="dashboard?.perf.cached">cache</span>
            <span v-else>ao vivo</span>
            <span v-if="dashboard?.perf.msTotal"> • {{ dashboard.perf.msTotal }}ms</span>
          </div>
        </div>

        <div class="mt-4 overflow-x-auto">
          <div class="min-w-[720px]">
            <div class="grid grid-cols-12 items-end gap-2">
              <div
                v-for="m in mensalChart"
                :key="m.key"
                class="flex flex-col items-center gap-2"
              >
                <div
                  class="w-full rounded-lg bg-emerald-500/80"
                  :style="{ height: m.heightPx + 'px' }"
                  :title="m.tooltip"
                />
                <div class="text-[11px] text-muted-foreground">{{ m.label }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-sm font-semibold">Top 10 Fornecedores</div>
          <div class="mt-3 space-y-2">
            <div v-if="(dashboard?.top.fornecedores.length ?? 0) === 0" class="text-sm text-muted-foreground">
              Sem dados.
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="i in dashboard!.top.fornecedores"
                :key="i.nome"
                class="flex items-center justify-between gap-3 rounded-xl bg-secondary px-3 py-2 text-sm"
              >
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ i.nome }}</div>
                  <div class="text-xs text-muted-foreground">{{ formatMoney(i.faturamento) }}</div>
                </div>
                <div class="shrink-0 text-xs font-semibold text-muted-foreground">{{ formatPct(i.percentual) }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-sm font-semibold">Top 10 Clientes</div>
          <div class="mt-3 space-y-2">
            <div v-if="(dashboard?.top.clientes.length ?? 0) === 0" class="text-sm text-muted-foreground">
              Sem dados.
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="i in dashboard!.top.clientes"
                :key="i.nome"
                class="flex items-center justify-between gap-3 rounded-xl bg-secondary px-3 py-2 text-sm"
              >
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ i.nome }}</div>
                  <div class="text-xs text-muted-foreground">{{ formatMoney(i.faturamento) }}</div>
                </div>
                <div class="shrink-0 text-xs font-semibold text-muted-foreground">{{ formatPct(i.percentual) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-sm font-semibold">Top 10 Produtos</div>
          <div class="mt-3 space-y-2">
            <div v-if="(dashboard?.top.produtos.length ?? 0) === 0" class="text-sm text-muted-foreground">
              Sem dados.
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="i in dashboard!.top.produtos"
                :key="i.nome"
                class="flex items-center justify-between gap-3 rounded-xl bg-secondary px-3 py-2 text-sm"
              >
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ i.nome }}</div>
                  <div class="text-xs text-muted-foreground">{{ formatMoney(i.faturamento) }}</div>
                </div>
                <div class="shrink-0 text-xs font-semibold text-muted-foreground">{{ formatPct(i.percentual) }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-sm font-semibold">Top 10 Vendedores</div>
          <div class="mt-3 space-y-2">
            <div v-if="(dashboard?.top.vendedores.length ?? 0) === 0" class="text-sm text-muted-foreground">
              Sem dados.
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="i in dashboard!.top.vendedores"
                :key="i.nome"
                class="flex items-center justify-between gap-3 rounded-xl bg-secondary px-3 py-2 text-sm"
              >
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ i.nome }}</div>
                  <div class="text-xs text-muted-foreground">{{ formatMoney(i.faturamento) }}</div>
                </div>
                <div class="shrink-0 text-xs font-semibold text-muted-foreground">{{ formatPct(i.percentual) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import AppShell from "../components/AppShell.vue";
import DashboardFilters from "../components/DashboardFilters.vue";
import { apiFetch } from "../lib/api";

type FilialOption = { id: number; label: string };

type DashboardTopItem = { nome: string; faturamento: number; percentual: number };

type VisaoGeralPayload = {
  filters: { dataInicio: string; dataFim: string; filiais: number[] };
  metrics: { faturamentoTotal: number; clientesAtivos: number; totalNFs: number; ticketMedio: number };
  mensal: { mes: string; faturamento: number }[];
  top: {
    fornecedores: DashboardTopItem[];
    clientes: DashboardTopItem[];
    produtos: DashboardTopItem[];
    vendedores: DashboardTopItem[];
  };
  perf: { cached: boolean; msTotal: number };
  warning?: string;
};

const filiais = ref<FilialOption[]>([]);
const filters = ref<{ dataInicio: string; dataFim: string; filiais: number[] }>({
  dataInicio: "",
  dataFim: "",
  filiais: [],
});

const loading = ref(false);
const error = ref<string>("");
const warning = ref<string>("");
const dashboard = ref<VisaoGeralPayload | null>(null);
let scheduled: number | null = null;

function isoDate(d: Date) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function formatMoney(v: number) {
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(v ?? 0);
}

function formatInt(v: number) {
  return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 }).format(v ?? 0);
}

function formatPct(v: number) {
  return new Intl.NumberFormat("pt-BR", { style: "percent", maximumFractionDigits: 1 }).format((v ?? 0) / 100);
}

const mensalChart = computed(() => {
  const items = dashboard.value?.mensal ?? [];
  const max = Math.max(1, ...items.map((x) => x.faturamento ?? 0));
  return items.map((x) => {
    const heightPx = Math.round(((x.faturamento ?? 0) / max) * 180);
    const d = new Date(x.mes + "T00:00:00");
    const label = new Intl.DateTimeFormat("pt-BR", { month: "2-digit", year: "2-digit" }).format(d);
    const tooltip = `${label} • ${formatMoney(x.faturamento ?? 0)}`;
    return { key: x.mes, label, tooltip, heightPx };
  });
});

async function loadFiliais() {
  const res = await apiFetch<{ filiais: FilialOption[] }>("/api/dashboard/filiais");
  if (!res.ok) {
    return;
  }
  filiais.value = res.data.filiais ?? [];
  if (filters.value.filiais.length === 0 && filiais.value.length > 0) {
    filters.value = { ...filters.value, filiais: filiais.value.map((x) => x.id) };
  }
}

async function reload() {
  loading.value = true;
  error.value = "";
  warning.value = "";
  try {
    const qs = new URLSearchParams();
    qs.set("data_inicio", filters.value.dataInicio);
    qs.set("data_final", filters.value.dataFim);
    if (filters.value.filiais.length > 0) {
      qs.set("filiais", filters.value.filiais.join(","));
    }

    const res = await apiFetch<VisaoGeralPayload>(`/api/dashboard/visao-geral?${qs.toString()}`);
    if (!res.ok) {
      error.value = res.error;
      dashboard.value = null;
      return;
    }
    dashboard.value = res.data;
    warning.value = res.data.warning ?? "";
  } finally {
    loading.value = false;
  }
}

function onFiltersChanged(p: { value: { dataInicio: string; dataFim: string; filiais: number[] }; valid: boolean }) {
  if (!p.valid) return;
  if (scheduled) window.clearTimeout(scheduled);
  scheduled = window.setTimeout(() => {
    void reload();
  }, 250);
}

onMounted(async () => {
  const today = new Date();
  const start = new Date(today);
  start.setMonth(today.getMonth() - 11);
  start.setDate(1);
  filters.value = { ...filters.value, dataInicio: isoDate(start), dataFim: isoDate(today) };
  await loadFiliais();
  await reload();
});
</script>
