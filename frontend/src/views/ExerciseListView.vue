<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const exercises = ref([])
const loading = ref(true)
const error = ref('')
const router = useRouter()

onMounted(async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_BASE_URL}/api/exercises`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    exercises.value = response.data
  } catch (err) {
    error.value = 'Failed to load exercises.'
    console.error(err)
  } finally {
    loading.value = false
  }
})

const startExercise = (id) => {
  router.push(`/exercises/${id}`)
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-violet-400">
        Coding Challenges
      </h1>
    </div>

    <div v-if="loading" class="text-gray-400 animate-pulse text-lg">Loading challenges...</div>
    <div v-else-if="error" class="text-red-400">{{ error }}</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="exercise in exercises" 
        :key="exercise.id"
        class="bg-gray-800/80 backdrop-blur border border-gray-700/50 rounded-2xl p-6 shadow-xl hover:shadow-2xl hover:-translate-y-1 transition-all duration-300 flex flex-col"
      >
        <div class="flex-grow">
          <div class="flex items-center space-x-3 mb-4">
            <div class="p-3 bg-indigo-500/20 text-indigo-400 rounded-xl">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="想定 path" />
                <path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
            </div>
            <h3 class="text-xl font-bold text-white">{{ exercise.title }}</h3>
          </div>
          <p class="text-gray-400 text-sm line-clamp-3 mb-6">
            {{ exercise.description }}
          </p>
        </div>
        
        <button 
          @click="startExercise(exercise.id)"
          class="w-full mt-auto py-3 px-4 bg-gradient-to-r from-indigo-500 to-violet-600 hover:from-indigo-400 hover:to-violet-500 text-white font-semibold rounded-xl shadow-lg hover:shadow-indigo-500/25 transition-all outline-none focus:ring-2 focus:ring-indigo-400 focus:ring-offset-2 focus:ring-offset-gray-900"
        >
          Solve Challenge
        </button>
      </div>
    </div>
  </div>
</template>
