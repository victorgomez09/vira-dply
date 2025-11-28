<script lang="ts" setup>
import { ref, onMounted, onServerPrefetch } from 'vue';

// --- Tipado de Datos ---
interface Project {
  ID: number;
  Name: string;
  Description: string;
  GitUrl: string;
  GitBranch: string;
  Status: 'Active' | 'Pending' | 'Completed';
  CreatedAt: string;
}

definePageMeta({
    middleware: ['auth']
})

// --- Estado Reactivo ---
// Usaremos ref para el estado local
const projects = ref<Project[] | null>(null);
const pending = ref(true);
const error = ref<any>(null);


// --- Función de Carga de Datos con $fetch ---
const fetchProjects = async () => {
  pending.value = true;
  error.value = null;

  try {
    // 🚨 Uso directo de $fetch. La URL es relativa para activar el proxy.
    const result = await $fetch<Project[]>('/api/projects');
    
    projects.value = result;
  } catch (e: any) {
    console.error("Fallo al cargar proyectos con $fetch:", e);
    error.value = e;
  } finally {
    pending.value = false;
  }
};

// --- Estrategia de Carga (Híbrida: SSR y CSR) ---

// 1. Carga en el Servidor (SSR): Solo si $fetch es accesible
if (process.server) {
  onServerPrefetch(fetchProjects);
}

// 2. Carga en el Cliente (CSR): Como fallback o para navegación
onMounted(() => {
  // Si los datos ya se cargaron en SSR, no hacemos nada.
  // Solo volvemos a cargar si no hay datos o si hubo un error en SSR.
  if (!projects.value && !error.value) {
    fetchProjects();
  }
});


// Función para obtener el color del status (se mantiene igual)
const getStatusColor = (status: Project['Status']) => {
  switch (status) {
    case 'Active': return 'success';
    case 'Pending': return 'warning';
    case 'Completed': return 'info';
    default: return 'neutral';
  }
}
</script>

<template>
  <div class="p-8">
    <h1 class="text-3xl font-bold mb-6">Mis Proyectos 📊</h1>

    <div v-if="pending" class="flex justify-center items-center h-48">
      <UProgress animation="swing" color="primary" />
      <p class="ml-4 text-gray-500 dark:text-gray-400">Cargando proyectos...</p>
    </div>

    <UAlert
      v-else-if="error"
      icon="i-heroicons-exclamation-triangle-20-solid"
      color="error"
      variant="soft"
      title="Error al cargar"
      :description="`No se pudieron cargar los proyectos. Error: ${error.message || error.statusText}`"
    />

    <div v-else-if="projects && projects.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <UCard variant="soft" v-for="project in projects" :key="project.ID" class="shadow-lg hover:shadow-xl transition duration-300">
        <template #header>
          <div class="flex justify-between items-center">
            <h2 class="text-xl font-semibold truncate">{{ project.Name }}</h2>
            <UBadge :color="getStatusColor(project.Status)" variant="subtle" size="sm">
              {{ project.Status }}
            </UBadge>
          </div>
        </template>

        <div class="flex flex-col">
            <p class="text-gray-600 dark:text-gray-300 line-clamp-3 mb-4">{{ project.Description }}</p>
            <div class="flex items-center gap-1">
                <UIcon name="i-pajamas-git" class="size-5" />
                <span>{{ project.GitUrl }}</span>
            </div>
        </div>

        <template #footer>
          <div class="flex justify-between items-center text-sm text-gray-400">
            <span>Creado: {{ new Date(project.CreatedAt).toLocaleDateString() }}</span>
          </div>
        </template>
      </UCard>
    </div>

    <div v-else class="text-center p-10 bg-gray-50 dark:bg-gray-800 rounded-lg">
      <UIcon name="i-heroicons-folder-open-20-solid" class="w-10 h-10 text-gray-400 mb-3" />
      <p class="text-lg text-gray-500 dark:text-gray-400">No hay proyectos disponibles todavía.</p>
    </div>
  </div>
</template>