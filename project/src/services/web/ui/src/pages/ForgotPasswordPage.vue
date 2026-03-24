<template>
  <div class="flex min-h-screen items-center justify-center px-4">
    <div class="w-full max-w-md space-y-6 rounded-2xl border border-border bg-card p-8 shadow-sm">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold tracking-tight">Recuperar senha</h1>
        <p class="text-sm text-muted-foreground">Enviaremos um link para redefinir.</p>
      </div>

      <div v-if="ctx.success" class="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
        {{ ctx.success }}
      </div>
      <div v-if="ctx.error" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
        {{ ctx.error }}
      </div>

      <form method="POST" action="/recuperar-senha" class="space-y-4">
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

        <input type="hidden" name="_csrf" :value="csrf" />

        <button
          type="submit"
          class="inline-flex w-full items-center justify-center rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
        >
          Enviar link
        </button>

        <div class="text-center">
          <RouterLink to="/login" class="text-sm font-medium text-primary hover:underline">Voltar</RouterLink>
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
</script>
