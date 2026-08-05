<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const login = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()

const handleLogin = async () => {
  error.value = ''
  loading.value = true
  try {
    const res = await axios.post(`${API_BASE_URL}/api/login`, {
      login: login.value,
      password: password.value
    })
    
    // Save token and user info
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    
    // Apply theme immediately
    if (res.data.user && res.data.user.theme) {
      localStorage.setItem('theme', res.data.user.theme)
      if (res.data.user.theme === 'light') {
        document.documentElement.classList.remove('dark')
      } else {
        document.documentElement.classList.add('dark')
      }
    }
    
    axios.defaults.headers.common['Authorization'] = `Bearer ${res.data.token}`
    
    // Redirect based on role
    router.push('/welcome')
    
  } catch (err) {
    if (err.response && err.response.data && err.response.data.error) {
      error.value = err.response.data.error
    } else {
      error.value = 'Failed to connect to the server.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-5xl bg-white dark:bg-slate-800 rounded-2xl shadow-xl overflow-hidden flex flex-col md:flex-row transform transition-all duration-300 card-lift">
    
    <!-- Left Side: Branding/Visual -->
    <div class="md:w-5/12 bg-gradient-to-br from-indigo-600 via-purple-600 to-violet-800 dark:from-indigo-800 dark:via-purple-800 dark:to-violet-900 p-10 flex flex-col justify-between text-white relative overflow-hidden transition-colors duration-300">
      <!-- Decorative background blur -->
      <div class="absolute -top-24 -left-24 w-64 h-64 bg-white/10 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-24 -right-24 w-80 h-80 bg-indigo-400/20 rounded-full blur-3xl"></div>
      
      <div class="relative z-10">
        <h2 class="text-4xl font-extrabold tracking-tight mb-4">Empowering<br/>The Future.</h2>
        <p class="text-indigo-100 text-lg">Join the most advanced programming education platform and unlock your full potential.</p>
      </div>

      <div class="relative z-10 mt-12 md:mt-0">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-md">
            <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
          </div>
          <span class="font-bold tracking-wider">Atom</span>
        </div>
      </div>
    </div>

    <!-- Right Side: Login Form -->
    <div class="md:w-7/12 p-8 sm:p-12 bg-white dark:bg-slate-800 transition-colors duration-300">
      <div class="mb-8">
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Welcome Back</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-2">Sign in to your account to continue</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-6">
        <div v-if="error" class="bg-red-50 border-l-4 border-red-500 p-4 rounded-md animate-pulse">
          <div class="flex">
            <div class="flex-shrink-0">
              <svg class="h-5 w-5 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
            </div>
            <div class="ml-3">
              <p class="text-sm text-red-700 font-medium">{{ error }}</p>
            </div>
          </div>
        </div>

        <div class="group">
          <label for="login" class="block text-sm font-semibold text-slate-700 dark:text-slate-300 transition-colors group-focus-within:text-indigo-600 dark:group-focus-within:text-indigo-400">Username or Email</label>
          <div class="mt-2 relative">
            <input id="login" v-model="login" type="text" required class="block w-full px-4 py-3 bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 dark:focus:border-indigo-400 focus:bg-white dark:focus:bg-slate-900 dark:text-white transition-all shadow-sm sm:text-sm outline-none" placeholder="user123 or user@example.com" />
          </div>
        </div>

        <div class="group">
          <label for="password" class="block text-sm font-semibold text-slate-700 dark:text-slate-300 transition-colors group-focus-within:text-indigo-600 dark:group-focus-within:text-indigo-400">Password</label>
          <div class="mt-2">
            <input id="password" v-model="password" type="password" required class="block w-full px-4 py-3 bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 dark:focus:border-indigo-400 focus:bg-white dark:focus:bg-slate-900 dark:text-white transition-all shadow-sm sm:text-sm outline-none" placeholder="••••••••" />
          </div>
        </div>

        <div class="pt-2">
          <button type="submit" :disabled="loading" class="w-full flex justify-center py-3.5 px-4 border border-transparent rounded-xl text-sm font-bold text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all duration-200 disabled:opacity-70 disabled:cursor-not-allowed transform active:scale-[0.98]">
            <span v-if="loading" class="flex items-center">
              <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Signing in...
            </span>
            <span v-else>Sign In</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
