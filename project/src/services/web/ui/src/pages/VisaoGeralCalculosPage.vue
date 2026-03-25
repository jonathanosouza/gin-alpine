<template>
  <AppShell>
    <div class="space-y-6">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Documentação de cálculos</h1>
          <p class="text-sm text-muted-foreground">
            Referência única das métricas e critérios usados na Visão Geral.
          </p>
        </div>
        <RouterLink
          to="/visao-geral"
          class="inline-flex items-center gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm hover:bg-blue-50"
        >
          Voltar
        </RouterLink>
      </div>

      <div class="rounded-2xl border border-border bg-card p-5">
        <div class="text-sm font-semibold">Filtros</div>
        <div class="mt-3 space-y-3 text-sm text-muted-foreground">
          <div>
            <div class="font-semibold text-foreground">Data inicial e data final</div>
            <div>O período sempre é aplicado a todas as métricas e rankings.</div>
          </div>
          <div>
            <div class="font-semibold text-foreground">Filiais (multi-seleção)</div>
            <div>
              Quando múltiplas filiais são selecionadas, os valores são somados automaticamente no backend.
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-border bg-card p-5">
        <div class="text-sm font-semibold">Métricas (cards)</div>
        <div class="mt-3 grid gap-4 md:grid-cols-2">
          <div class="rounded-xl bg-secondary p-4">
            <div class="text-sm font-semibold">Faturamento Total</div>
            <div class="mt-1 text-sm text-muted-foreground">
              Soma do valor total das notas fiscais no período e filiais selecionadas.
            </div>
            <div class="mt-2 text-xs text-muted-foreground">
              Cálculo: Faturamento Total = Σ(VL_TOTAL)
            </div>
          </div>
          <div class="rounded-xl bg-secondary p-4">
            <div class="text-sm font-semibold">Quantidade de Clientes Ativos</div>
            <div class="mt-1 text-sm text-muted-foreground">
              Contagem de clientes distintos com pelo menos uma nota fiscal no período.
            </div>
            <div class="mt-2 text-xs text-muted-foreground">
              Cálculo: Clientes Ativos = COUNT(DISTINCT COD_CLIENTE)
            </div>
          </div>
          <div class="rounded-xl bg-secondary p-4">
            <div class="text-sm font-semibold">Quantidade Total de Notas Fiscais</div>
            <div class="mt-1 text-sm text-muted-foreground">
              Total de notas fiscais contabilizadas no período e filiais selecionadas.
            </div>
            <div class="mt-2 text-xs text-muted-foreground">
              Cálculo: Total NFs = COUNT(*)
            </div>
          </div>
          <div class="rounded-xl bg-secondary p-4">
            <div class="text-sm font-semibold">Ticket Médio</div>
            <div class="mt-1 text-sm text-muted-foreground">
              Média de faturamento por nota fiscal.
            </div>
            <div class="mt-2 text-xs text-muted-foreground">
              Cálculo: Ticket Médio = Faturamento Total ÷ Total NFs
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-border bg-card p-5">
        <div class="text-sm font-semibold">Gráfico mês a mês</div>
        <div class="mt-2 text-sm text-muted-foreground">
          Agrega o faturamento por mês (date_trunc('month')) sempre respeitando os filtros.
        </div>
      </div>

      <div class="rounded-2xl border border-border bg-card p-5">
        <div class="text-sm font-semibold">Rankings (Top 10)</div>
        <div class="mt-3 space-y-3 text-sm text-muted-foreground">
          <div>
            <div class="font-semibold text-foreground">Percentual</div>
            <div>
              Em todos os rankings: Percentual = (Faturamento do item ÷ Faturamento Total do período) × 100.
            </div>
          </div>
          <div>
            <div class="font-semibold text-foreground">Top 10 Fornecedores / Produtos</div>
            <div>
              Baseado no faturamento agregado por fornecedor/produto dentro do período e filiais selecionadas.
            </div>
          </div>
          <div>
            <div class="font-semibold text-foreground">Top 10 Clientes / Vendedores</div>
            <div>
              Baseado no faturamento agregado por cliente/vendedor dentro do período e filiais selecionadas.
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router";
import AppShell from "../components/AppShell.vue";
</script>
