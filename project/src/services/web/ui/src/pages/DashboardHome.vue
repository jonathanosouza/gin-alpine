<template>
  <AppShell>
    <div class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
          <p class="text-sm text-muted-foreground">Área do BI (as telas entram aqui).</p>
        </div>
        <div class="flex items-center gap-2">
          <div class="rounded-xl border border-border bg-card px-3 py-2 text-sm">
            Schema: <span class="font-semibold">agrocel</span>
          </div>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total pedidos</div>
          <div class="mt-2 text-2xl font-bold">—</div>
          <div class="mt-1 text-xs text-muted-foreground">Aguardando integração do BI</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Faturamento</div>
          <div class="mt-2 text-2xl font-bold">—</div>
          <div class="mt-1 text-xs text-muted-foreground">Aguardando integração do BI</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Clientes</div>
          <div class="mt-2 text-2xl font-bold">—</div>
          <div class="mt-1 text-xs text-muted-foreground">Aguardando integração do BI</div>
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Produtos</div>
          <div class="mt-2 text-2xl font-bold">—</div>
          <div class="mt-1 text-xs text-muted-foreground">Aguardando integração do BI</div>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-3">
        <div class="rounded-2xl border border-border bg-card p-4 lg:col-span-2">
          <div class="flex items-center justify-between">
            <div class="text-sm font-semibold">Espaço para gráfico</div>
            <div class="text-xs text-muted-foreground">Em breve</div>
          </div>
          <div class="mt-4 h-64 rounded-xl bg-secondary" />
        </div>
        <div class="rounded-2xl border border-border bg-card p-4">
          <div class="text-sm font-semibold">Notas</div>
          <div class="mt-4 space-y-2 text-sm text-muted-foreground">
            <div class="rounded-xl bg-secondary px-3 py-2">Conectar queries do BI aqui.</div>
            <div class="rounded-xl bg-secondary px-3 py-2">Adicionar cards e filtros.</div>
            <div class="rounded-xl bg-secondary px-3 py-2">Ativar export e drilldown.</div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import AppShell from "../components/AppShell.vue";
</script>
