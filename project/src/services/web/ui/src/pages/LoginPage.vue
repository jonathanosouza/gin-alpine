<template>
  <div class="flex min-h-screen items-center justify-center px-4">
    <div class="w-full max-w-md space-y-6 rounded-2xl border border-border bg-card p-8 shadow-sm">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold tracking-tight">Entrar</h1>
        <p class="text-sm text-muted-foreground">Acesse o painel do BI.</p>
      </div>

      <div v-if="ctx.error" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
        {{ ctx.error }}
      </div>

      <form method="POST" :action="loginAction" class="space-y-4">
        <div class="space-y-2">
          <label class="text-sm font-medium">Email</label>
          <input
            v-model="email"
            name="email"
            type="email"
            required
            class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
            autocomplete="email"
          />
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium">Senha</label>
          <input
            v-model="password"
            name="password"
            type="password"
            required
            class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
            autocomplete="current-password"
          />
        </div>

        <input type="hidden" name="_csrf" :value="csrf" />

        <button
          type="submit"
          class="inline-flex w-full items-center justify-center rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
        >
          Entrar
        </button>

        <div class="text-center">
          <RouterLink to="/recuperar-senha" class="text-sm font-medium text-primary hover:underline">
            Esqueci minha senha
          </RouterLink>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";
import { getAppContext, getCSRFToken } from "../lib/appContext";

const ctx = getAppContext();
const csrf = getCSRFToken();
const email = ref("");
const password = ref("");
const loginAction =
  window.location.port === "3000" ? "/login" : "http://localhost:3000/login";
</script>
