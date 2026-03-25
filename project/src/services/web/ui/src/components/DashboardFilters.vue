<template>
  <div class="rounded-2xl border border-border bg-card p-4">
    <div class="flex flex-wrap items-end gap-4">
      <div class="min-w-[220px]">
        <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Data inicial</div>
        <input
          :value="modelValue.dataInicio"
          type="date"
          class="mt-1 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm"
          @change="onDataInicioChange"
        />
      </div>

      <div class="min-w-[220px]">
        <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Data final</div>
        <input
          :value="modelValue.dataFim"
          type="date"
          class="mt-1 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm"
          @change="onDataFimChange"
        />
      </div>

      <div class="flex-1" />

      <button
        type="button"
        class="inline-flex items-center justify-center rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
        :disabled="loading"
        @click="onSubmit"
      >
        Atualizar
      </button>
    </div>

    <div class="mt-4">
      <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Filiais</div>
      <div v-if="filiaisOptions.length === 0" class="mt-2 text-sm text-muted-foreground">
        Nenhuma filial disponível.
      </div>
      <div v-else class="mt-2 grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
        <label
          v-for="f in filiaisOptions"
          :key="f.id"
          class="flex items-center gap-2 rounded-xl border border-border bg-background px-3 py-2 text-sm"
        >
          <input
            type="checkbox"
            class="h-4 w-4"
            :value="f.id"
            :checked="modelValue.filiais.includes(f.id)"
            @change="onFilialToggle(f.id, ($event.target as HTMLInputElement).checked)"
          />
          <span class="truncate">{{ f.label }}</span>
        </label>
      </div>
    </div>

    <div v-if="localError" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {{ localError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

type FilialOption = { id: number; label: string };

type FiltersModel = {
  dataInicio: string;
  dataFim: string;
  filiais: number[];
};

const props = defineProps<{
  modelValue: FiltersModel;
  filiaisOptions: FilialOption[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: FiltersModel): void;
  (e: "change", p: { value: FiltersModel; valid: boolean; error: string }): void;
  (e: "submit", v: FiltersModel): void;
}>();

const localError = ref("");

function validate(v: FiltersModel) {
  if (!v.dataInicio || !v.dataFim) {
    return "Informe data inicial e data final.";
  }
  const a = new Date(v.dataInicio + "T00:00:00");
  const b = new Date(v.dataFim + "T00:00:00");
  if (Number.isNaN(a.getTime()) || Number.isNaN(b.getTime())) {
    return "Formato de data inválido.";
  }
  if (a.getTime() > b.getTime()) {
    return "Data inicial não pode ser maior que a data final.";
  }
  const maxDays = 366 * 3;
  const diffDays = Math.floor((b.getTime() - a.getTime()) / (24 * 3600 * 1000));
  if (diffDays > maxDays) {
    return "Intervalo muito grande. Selecione até 3 anos.";
  }
  return "";
}

function emitChange(v: FiltersModel) {
  const err = validate(v);
  localError.value = err;
  emit("change", { value: v, valid: err === "", error: err });
}

function onDataInicioChange(e: Event) {
  const value = (e.target as HTMLInputElement).value;
  const next = { ...props.modelValue, dataInicio: value };
  emit("update:modelValue", next);
  emitChange(next);
}

function onDataFimChange(e: Event) {
  const value = (e.target as HTMLInputElement).value;
  const next = { ...props.modelValue, dataFim: value };
  emit("update:modelValue", next);
  emitChange(next);
}

function onFilialToggle(id: number, checked: boolean) {
  const current = props.modelValue.filiais ?? [];
  const set = new Set<number>(current);
  if (checked) set.add(id);
  else set.delete(id);
  const next = { ...props.modelValue, filiais: Array.from(set).sort((a, b) => a - b) };
  emit("update:modelValue", next);
  emitChange(next);
}

function onSubmit() {
  const err = validate(props.modelValue);
  localError.value = err;
  if (err) return;
  emit("submit", props.modelValue);
}
</script>
