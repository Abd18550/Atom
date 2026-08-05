<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { API_BASE_URL } from './config.js'

const route = useRoute()
const router = useRouter()

const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))

// Watch for changes in the route to ensure the user object is up-to-date
// (e.g. after a login or profile update)
watch(
  () => route.path,
  () => {
    user.value = JSON.parse(localStorage.getItem('user') || '{}')
  }
)

const showNav = computed(() => route.path !== '/login')
const isWelcomePage = computed(() => route.path === '/welcome' || route.path === '/student-home')

const goBack = () => {
  router.back()
}

const goToProfile = () => {
  router.push('/profile')
}

const signOut = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  user.value = {}
  router.push('/login')
}

// Dark Mode logic
const isDark = ref(true)

const initTheme = () => {
  // If we have a logged-in user with a theme, prioritize that
  let savedTheme = localStorage.getItem('theme')
  
  if (user.value && user.value.theme) {
    savedTheme = user.value.theme
    localStorage.setItem('theme', savedTheme)
  }

  if (savedTheme === 'light') {
    isDark.value = false
  } else {
    isDark.value = true // Default to dark mode
  }
  applyTheme()
}

const applyTheme = () => {
  if (isDark.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

const toggleDark = async () => {
  isDark.value = !isDark.value
  const themeString = isDark.value ? 'dark' : 'light'
  
  localStorage.setItem('theme', themeString)
  applyTheme()
  
  if (user.value && user.value.id) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/theme`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ theme: themeString })
      });
      
      if (response.ok) {
        user.value.theme = themeString;
        localStorage.setItem('user', JSON.stringify(user.value));
      } else {
        console.error("Failed to update theme on server")
      }
    } catch (err) {
      console.error("Error updating theme:", err)
    }
  }
}

onMounted(() => {
  initTheme()
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-900 flex flex-col font-sans transition-colors duration-300">
    <header class="sticky top-0 z-50 glass dark:bg-slate-900/70 dark:border-white/10 transition-colors duration-300">
      <div class="max-w-7xl mx-auto flex justify-between items-center py-4 px-6 md:px-12 text-slate-800 dark:text-slate-200">
        <!-- Left side: Back button (if authenticated AND not on Welcome page) plus Title -->
        <div class="flex items-center space-x-4">
          <template v-if="showNav">
            <button v-if="!isWelcomePage" @click="goBack" class="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors" title="Go Back">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
            </button>
          </template>
          <router-link :to="user && user.role === 'Student' ? '/student-home' : '/welcome'" class="text-4xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-indigo-800 hover:opacity-80 transition-opacity">Atom</router-link>
        </div>
        
        <!-- Right side: Profile and Sign Out button (if authenticated) -->
        <div v-if="showNav" class="flex items-center space-x-2 sm:space-x-4">
            <button @click="toggleDark" class="flex items-center justify-center p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors" title="Toggle Dark Mode">
              <svg v-if="isDark" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path></svg>
            </button>
            <button @click="goToProfile" class="flex items-center p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors group" title="My Profile">
                <img v-if="user && user.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + user.avatar" class="w-6 h-6 rounded-full bg-indigo-100 dark:bg-indigo-900 group-hover:scale-110 transition-transform mr-1" />
                <svg v-else class="w-5 h-5 group-hover:scale-110 transition-transform mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
                <span v-if="user && user.username" class="text-sm font-bold tracking-wide pr-1 dark:text-slate-200">
                  {{ user.username }}
                </span>
            </button>
          <button @click="signOut" class="flex items-center space-x-2 bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 hover:bg-indigo-100 dark:hover:bg-indigo-500/30 px-4 py-2 rounded-lg transition-colors text-sm font-medium text-indigo-700 dark:text-indigo-300">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path></svg>
            <span class="hidden sm:inline">Sign Out</span>
          </button>
        </div>
      </div>
    </header>

    <main class="flex-grow flex items-center justify-center p-6">
      <RouterView />
    </main>

    <footer class="bg-gray-100 dark:bg-slate-800 py-4 text-center text-sm text-gray-500 dark:text-slate-400 border-t border-gray-200 dark:border-slate-700 transition-colors duration-300">
      &copy; 2026 Programming Education Platform
    </footer>
  </div>
</template>

<style>
/* App global overrides if necessary */
</style>
