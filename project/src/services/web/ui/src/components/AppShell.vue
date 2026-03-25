<template>
  <div class="min-h-screen bg-background text-foreground">
    <div class="flex min-h-screen">
      <div
        v-show="sidebarOpen"
        class="fixed inset-0 z-30 bg-black/40 lg:hidden"
        @click="sidebarOpen = false"
      />

      <aside
        class="sidebar fixed inset-y-0 left-0 z-40 w-72 -translate-x-full border-r border-border bg-card transition-transform lg:static lg:translate-x-0"
        :class="sidebarOpen ? 'translate-x-0' : ''"
      >
        <div class="flex h-16 items-center justify-center border-b border-border px-2">
          <RouterLink to="/" class="inline-flex items-center">
            <img :src="logoSrc" alt="Logo" class="h-10 w-auto object-contain" style="filter:none;mix-blend-mode:normal" />
          </RouterLink>
        </div>

        <div class="px-3 py-4">
          <nav class="space-y-2">
            <RouterLink to="/visao-geral" class="block rounded-lg px-3 py-2 text-sm font-semibold">
              Visão Geral
            </RouterLink>
            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-sm font-semibold"
              @click="openComercial = !openComercial"
            >
              <span class="flex items-center gap-3">
                <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3 3v18h18" />
                  <path d="M13 17V9" />
                  <path d="M18 17V5" />
                  <path d="M8 17v-3" />
                </svg>
                Comercial
              </span>
              <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path :d="openComercial ? 'M6 15l6-6 6 6' : 'M6 9l6 6 6-6'" />
              </svg>
            </button>
            <div v-show="openComercial" class="space-y-1 pl-9">
              <RouterLink
                to="/comercial/vendedor"
                class="block rounded-lg px-3 py-2 text-sm"
              >
                Vendas Vendedor
              </RouterLink>
              <RouterLink to="/comercial/cliente" class="block rounded-lg px-3 py-2 text-sm">Vendas Cliente</RouterLink>
              <RouterLink to="/comercial/fornecedor" class="block rounded-lg px-3 py-2 text-sm">Vendas Fornecedor</RouterLink>
              <RouterLink to="/comercial/real-meta" class="block rounded-lg px-3 py-2 text-sm">Real x Meta</RouterLink>
              <RouterLink to="/comercial/vendas-periodo" class="block rounded-lg px-3 py-2 text-sm">Vendas por Período</RouterLink>
              <RouterLink to="/comercial/crescimento-ano" class="block rounded-lg px-3 py-2 text-sm">Crescimento Ano</RouterLink>
              <RouterLink to="/comercial/evolucao-vendedor" class="block rounded-lg px-3 py-2 text-sm">Evolução Vendedor</RouterLink>
            </div>

            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-sm font-semibold"
              @click="openLogistica = !openLogistica"
            >
              <span class="flex items-center gap-3">
                <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3 9h18" />
                  <path d="M3 15h18" />
                  <path d="M6 22V2" />
                </svg>
                Logística
              </span>
              <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path :d="openLogistica ? 'M6 15l6-6 6 6' : 'M6 9l6 6 6-6'" />
              </svg>
            </button>
            <div v-show="openLogistica" class="space-y-1 pl-9">
              <RouterLink to="/logistica/estoque" class="block rounded-lg px-3 py-2 text-sm">Estoque</RouterLink>
              <RouterLink to="/logistica/sugestao-compras" class="block rounded-lg px-3 py-2 text-sm">Sugestão de Compras</RouterLink>
              <RouterLink to="/logistica/curva-abc" class="block rounded-lg px-3 py-2 text-sm">Curva ABC</RouterLink>
            </div>

            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-sm font-semibold"
              @click="openFinanceiro = !openFinanceiro"
            >
              <span class="flex items-center gap-3">
                <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="9" />
                  <path d="M12 7v10" />
                  <path d="M8 10h8" />
                </svg>
                Financeiro
              </span>
              <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path :d="openFinanceiro ? 'M6 15l6-6 6 6' : 'M6 9l6 6 6-6'" />
              </svg>
            </button>
            <div v-show="openFinanceiro" class="space-y-1 pl-9">
              <RouterLink to="/financeiro/contas-pagar" class="block rounded-lg px-3 py-2 text-sm">Contas a Pagar</RouterLink>
              <RouterLink to="/financeiro/contas-receber" class="block rounded-lg px-3 py-2 text-sm">Contas a Receber</RouterLink>
              <RouterLink to="/financeiro/fluxo-caixa" class="block rounded-lg px-3 py-2 text-sm">Fluxo de Caixa</RouterLink>
              <RouterLink to="/financeiro/dre" class="block rounded-lg px-3 py-2 text-sm">DRE</RouterLink>
            </div>

            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-sm font-semibold"
              @click="openConfig = !openConfig"
            >
              <span class="flex items-center gap-3">
                <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 2v4" />
                  <path d="M12 18v4" />
                  <path d="M4.93 4.93l2.83 2.83" />
                  <path d="M16.24 16.24l2.83 2.83" />
                  <path d="M2 12h4" />
                  <path d="M18 12h4" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
                Configuração
              </span>
              <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path :d="openConfig ? 'M6 15l6-6 6 6' : 'M6 9l6 6 6-6'" />
              </svg>
            </button>
            <div v-show="openConfig" class="space-y-1 pl-9">
              <RouterLink
                to="/configuracao"
                class="block rounded-lg px-3 py-2 text-sm"
              >
                Geral
              </RouterLink>
              <RouterLink to="/configuracao/usuarios" class="block rounded-lg px-3 py-2 text-sm">Usuários</RouterLink>
              <RouterLink to="/configuracao/metas" class="block rounded-lg px-3 py-2 text-sm">Cadastrar Metas</RouterLink>
              <RouterLink to="/configuracao/automacao" class="block rounded-lg px-3 py-2 text-sm">Automação</RouterLink>
              <RouterLink to="/perfil" class="block rounded-lg px-3 py-2 text-sm">Minha Senha</RouterLink>
            </div>
          </nav>
        </div>
      </aside>

      <div class="flex min-w-0 flex-1 flex-col">
        <header class="sticky top-0 z-20 border-b border-border bg-background/90 backdrop-blur">
          <div class="flex h-16 items-center gap-3 px-4 lg:px-6">
            <button
              type="button"
              class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-border hover:bg-blue-50 lg:hidden"
              @click="sidebarOpen = true"
            >
              <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 12h18" />
                <path d="M3 6h18" />
                <path d="M3 18h18" />
              </svg>
            </button>

            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div class="flex items-center gap-2">
                <select class="rounded-xl border border-input bg-background px-3 py-2 text-sm">
                  <option>Golamp Tecnologia</option>
                </select>
                <button type="button" class="inline-flex items-center gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm hover:bg-blue-50">
                  <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M8 3h8v4H8z" />
                    <path d="M4 7h16v14H4z" />
                  </svg>
                  Modo Apresentação
                </button>
              </div>
            </div>

            <div class="flex items-center gap-3">
              <button
                type="button"
                class="relative inline-flex h-10 w-10 items-center justify-center rounded-lg border border-border hover:bg-blue-50"
              >
                <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
                  <path d="M13.73 21a2 2 0 0 1-3.46 0" />
                </svg>
              </button>

              <div class="relative">
                <button
                  type="button"
                  class="flex items-center gap-3 rounded-xl border border-border bg-card px-3 py-2 hover:bg-blue-50"
                  @click="userMenuOpen = !userMenuOpen"
                >
                  <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary/15 text-primary">
                    <span class="text-sm font-bold">{{ initials }}</span>
                  </div>
                  <div class="hidden text-left sm:block">
                    <div class="max-w-[180px] truncate text-sm font-medium">{{ ctx.User?.Name ?? "-" }}</div>
                    <div class="max-w-[180px] truncate text-xs text-muted-foreground">{{ ctx.User?.Email ?? "" }}</div>
                  </div>
                  <svg class="h-4 w-4 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M6 9l6 6 6-6" />
                  </svg>
                </button>

                <div
                  v-show="userMenuOpen"
                  class="absolute right-0 mt-2 w-48 rounded-xl border border-border bg-card p-1 shadow-md"
                >
                  <RouterLink
                    to="/perfil"
                    class="block rounded-lg px-3 py-2 text-sm hover:bg-blue-50"
                    @click="userMenuOpen = false"
                  >
                    Meu perfil
                  </RouterLink>
                  <button
                    type="button"
                    class="w-full rounded-lg px-3 py-2 text-left text-sm hover:bg-blue-50"
                    @click="logout"
                  >
                    Sair
                  </button>
                </div>
              </div>
            </div>
          </div>
        </header>

        <main class="min-w-0 flex-1 px-4 py-6 lg:px-6">
          <slot />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { getAppContext, getCSRFToken } from "../lib/appContext";

const route = useRoute();
const ctx = getAppContext();
const sidebarOpen = ref(false);
const userMenuOpen = ref(false);
const openComercial = ref(true);
const openLogistica = ref(false);
const openFinanceiro = ref(false);
const openConfig = ref(true);
const logoSrc = "/logo4.png?v=20260324";

function isActive(path: string) {
  return route.path === path;
}

const initials = computed(() => {
  const name = String(ctx.User?.Name ?? "").trim();
  if (!name) return "U";
  const parts = name.split(/\s+/).filter(Boolean);
  const first = parts[0]?.[0] ?? "U";
  const last = parts.length > 1 ? parts[parts.length - 1]?.[0] ?? "" : "";
  return (first + last).toUpperCase();
});

async function logout() {
  const csrf = getCSRFToken();
  const base = window.location.port === "3000" ? "" : "http://localhost:3000";
  await fetch(base + "/logout", {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
    },
    credentials: "include",
  }).catch(() => undefined);
  window.location.href = window.location.pathname.startsWith("/static/vue/") ? "/static/vue/login" : "/login";
}
</script>
<style scoped>
.sidebar nav button:hover,
.sidebar nav a:hover {
  background-color: var(--color-muted);
}
.sidebar nav a.router-link-active { }
.sidebar nav a.router-link-active {
  /* no active color; keep default */
}
</style>
