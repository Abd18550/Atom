<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const router = useRouter()
const platformStats = ref({ totalStudents: 0, totalGroups: 0, totalStages: 0, totalExercises: 0 })
const loading = ref(true)

onMounted(async () => {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }

  try {
    const token = localStorage.getItem('token')
    const headers = { Authorization: `Bearer ${token}` }

    // Fetch stats in parallel
    const promises = []

    if (user.value.role === 'Admin' || user.value.role === 'Supervisor') {
      promises.push(
        axios.get(`${API_BASE_URL}/api/users`, { headers }).then(r => {
          platformStats.value.totalStudents = r.data.filter(u => u.role === 'Student').length
        }).catch(() => {}),
        axios.get(`${API_BASE_URL}/api/groups`, { headers }).then(r => {
          platformStats.value.totalGroups = r.data.length
        }).catch(() => {}),
        axios.get(`${API_BASE_URL}/api/stages`, { headers }).then(r => {
          platformStats.value.totalStages = r.data.length
          platformStats.value.totalExercises = r.data.reduce((acc, s) => acc + (s.questions?.length || 0), 0)
        }).catch(() => {})
      )
    } else if (user.value.role === 'Mentor') {
      promises.push(
        axios.get(`${API_BASE_URL}/api/groups`, { headers }).then(r => {
          platformStats.value.totalGroups = r.data.length
        }).catch(() => {})
      )
    }

    await Promise.all(promises)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
})

const greetingMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
})

const roleLabel = computed(() => {
  const labels = { Admin: 'Administrator', Supervisor: 'Supervisor', Mentor: 'Mentor' }
  return labels[user.value.role] || user.value.role
})
</script>

<template>
  <div class="w-full max-w-5xl space-y-6 animate-fade-in-up">

    <!-- Hero Banner -->
    <div class="relative bg-gradient-to-br from-slate-800 via-slate-900 to-slate-950 dark:from-slate-800 dark:via-slate-900 dark:to-black rounded-3xl overflow-hidden shadow-2xl shadow-slate-900/50 dark:shadow-black/50 border border-slate-700/50">
      <!-- Geometric accent shapes -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute top-[-30%] right-[-10%] w-80 h-80 bg-indigo-500/10 rounded-full blur-3xl"></div>
        <div class="absolute bottom-[-20%] left-[-10%] w-64 h-64 bg-purple-500/10 rounded-full blur-3xl"></div>
        <div class="absolute top-[50%] right-[30%] w-40 h-40 bg-cyan-500/5 rounded-full blur-2xl"></div>
        <!-- Grid pattern overlay -->
        <div class="absolute inset-0" style="background-image: radial-gradient(rgba(255,255,255,0.03) 1px, transparent 1px); background-size: 24px 24px;"></div>
      </div>

      <div class="relative z-10 p-8 md:p-10">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
          <!-- Left: Greeting -->
          <div>
            <div class="flex items-center gap-2 mb-3">
              <span class="px-3 py-1 text-xs font-bold uppercase tracking-wider rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                {{ roleLabel }}
              </span>
            </div>
            <p class="text-slate-400 text-sm font-semibold uppercase tracking-widest mb-1">{{ greetingMessage }}</p>
            <h1 class="text-3xl md:text-4xl font-black text-white tracking-tight">
              {{ user.full_name || user.username }}
            </h1>
            <p class="text-slate-400 text-sm mt-2">Welcome to your Atom dashboard.</p>
          </div>

          <!-- Right: Date / Time Badge -->
          <div class="hidden md:flex items-center gap-3">
            <div class="text-right">
              <p class="text-slate-300 text-sm font-bold">{{ new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' }) }}</p>
              <p class="text-slate-500 text-xs mt-0.5">Atom Platform</p>
            </div>
            <div class="w-12 h-12 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center">
              <svg class="w-6 h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats Overview (Admin/Supervisor) -->
    <div v-if="(user.role === 'Admin' || user.role === 'Supervisor') && !loading" class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
      <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 group">
        <div class="p-2.5 bg-indigo-100 dark:bg-indigo-500/15 rounded-xl text-indigo-600 dark:text-indigo-400 w-fit mb-3 group-hover:scale-110 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
        </div>
        <p class="text-2xl font-black text-slate-900 dark:text-white">{{ platformStats.totalStudents }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Students</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 group">
        <div class="p-2.5 bg-teal-100 dark:bg-teal-500/15 rounded-xl text-teal-600 dark:text-teal-400 w-fit mb-3 group-hover:scale-110 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
        </div>
        <p class="text-2xl font-black text-slate-900 dark:text-white">{{ platformStats.totalGroups }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Groups</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 group">
        <div class="p-2.5 bg-emerald-100 dark:bg-emerald-500/15 rounded-xl text-emerald-600 dark:text-emerald-400 w-fit mb-3 group-hover:scale-110 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16"></path></svg>
        </div>
        <p class="text-2xl font-black text-slate-900 dark:text-white">{{ platformStats.totalStages }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Stages</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 group">
        <div class="p-2.5 bg-purple-100 dark:bg-purple-500/15 rounded-xl text-purple-600 dark:text-purple-400 w-fit mb-3 group-hover:scale-110 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        </div>
        <p class="text-2xl font-black text-slate-900 dark:text-white">{{ platformStats.totalExercises }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Questions</p>
      </div>
    </div>

    <!-- Management Actions Section -->
    <div>
      <h3 class="text-sm font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-4 px-1">Management</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

        <!-- User Management (Admin/Supervisor) -->
        <router-link v-if="user.role === 'Admin' || user.role === 'Supervisor'" to="/admin" class="group flex items-center p-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
          <div class="absolute right-0 top-0 w-24 h-24 bg-indigo-50 dark:bg-indigo-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 bg-gradient-to-br from-indigo-500 to-indigo-600 dark:from-indigo-600 dark:to-indigo-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-indigo-200 dark:group-hover:shadow-indigo-900/50 group-hover:scale-110 transition-all">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
          </div>
          <div class="relative z-10 ml-5">
            <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">User Management</p>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Manage accounts</p>
          </div>
          <div class="relative z-10 ml-auto text-slate-300 dark:text-slate-600 group-hover:text-indigo-500 group-hover:translate-x-1 transition-all mr-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </router-link>

        <!-- Manage Stages (Admin/Supervisor) -->
        <router-link v-if="user.role === 'Admin' || user.role === 'Supervisor'" to="/admin/stages" class="group flex items-center p-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
          <div class="absolute right-0 top-0 w-24 h-24 bg-emerald-50 dark:bg-emerald-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 bg-gradient-to-br from-emerald-500 to-emerald-600 dark:from-emerald-600 dark:to-emerald-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-emerald-200 dark:group-hover:shadow-emerald-900/50 group-hover:scale-110 transition-all">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16"></path></svg>
          </div>
          <div class="relative z-10 ml-5">
            <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">Manage Stages</p>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Edit Learning Path nodes</p>
          </div>
          <div class="relative z-10 ml-auto text-slate-300 dark:text-slate-600 group-hover:text-emerald-500 group-hover:translate-x-1 transition-all mr-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </router-link>

        <!-- Manage Challenges (Admin/Supervisor) -->
        <router-link v-if="user.role === 'Admin' || user.role === 'Supervisor'" to="/admin/exercises" class="group flex items-center p-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
          <div class="absolute right-0 top-0 w-24 h-24 bg-pink-50 dark:bg-pink-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 bg-gradient-to-br from-pink-500 to-pink-600 dark:from-pink-600 dark:to-pink-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-pink-200 dark:group-hover:shadow-pink-900/50 group-hover:scale-110 transition-all">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg>
          </div>
          <div class="relative z-10 ml-5">
            <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-pink-600 dark:group-hover:text-pink-400 transition-colors">Manage Challenges</p>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Create & edit test cases</p>
          </div>
          <div class="relative z-10 ml-auto text-slate-300 dark:text-slate-600 group-hover:text-pink-500 group-hover:translate-x-1 transition-all mr-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </router-link>

        <!-- Mentor Panel (Mentor / Admin / Supervisor) -->
        <router-link v-if="user.role === 'Mentor' || user.role === 'Admin' || user.role === 'Supervisor'" to="/mentor" class="group flex items-center p-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
          <div class="absolute right-0 top-0 w-24 h-24 bg-blue-50 dark:bg-blue-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 bg-gradient-to-br from-blue-500 to-blue-600 dark:from-blue-600 dark:to-blue-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-blue-200 dark:group-hover:shadow-blue-900/50 group-hover:scale-110 transition-all">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
          </div>
          <div class="relative z-10 ml-5">
            <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">Mentor Panel</p>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Manage students</p>
          </div>
          <div class="relative z-10 ml-auto text-slate-300 dark:text-slate-600 group-hover:text-blue-500 group-hover:translate-x-1 transition-all mr-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </router-link>

        <!-- Student Groups -->
        <router-link to="/groups" class="group flex items-center p-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-100 dark:border-slate-700/50 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
          <div class="absolute right-0 top-0 w-24 h-24 bg-teal-50 dark:bg-teal-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 bg-gradient-to-br from-teal-500 to-teal-600 dark:from-teal-600 dark:to-teal-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-teal-200 dark:group-hover:shadow-teal-900/50 group-hover:scale-110 transition-all">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
          </div>
          <div class="relative z-10 ml-5">
            <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-teal-600 dark:group-hover:text-teal-400 transition-colors">Student Groups</p>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Manage classes</p>
          </div>
          <div class="relative z-10 ml-auto text-slate-300 dark:text-slate-600 group-hover:text-teal-500 group-hover:translate-x-1 transition-all mr-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </router-link>
      </div>
    </div>

    <!-- Open as Student Section -->
    <div>
      <h3 class="text-sm font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-4 px-1">Student View</h3>
      <router-link to="/student-home" class="group flex items-center p-5 bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-500/5 dark:to-orange-500/5 rounded-2xl border border-amber-200/50 dark:border-amber-500/20 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden relative">
        <div class="absolute right-0 top-0 w-32 h-32 bg-amber-100/50 dark:bg-amber-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
        <div class="relative z-10 bg-gradient-to-br from-amber-500 to-orange-600 dark:from-amber-600 dark:to-orange-700 text-white p-3.5 rounded-xl shadow-md group-hover:shadow-amber-200 dark:group-hover:shadow-amber-900/50 group-hover:scale-110 transition-all">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>
        </div>
        <div class="relative z-10 ml-5 flex-1">
          <p class="text-base font-bold text-slate-800 dark:text-slate-200 group-hover:text-amber-600 dark:group-hover:text-amber-400 transition-colors">Open as Student</p>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Preview the student dashboard, learning path & XP system</p>
        </div>
        <div class="relative z-10 ml-auto text-amber-400 dark:text-amber-500/50 group-hover:text-amber-600 group-hover:translate-x-1 transition-all mr-2">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
        </div>
      </router-link>
    </div>

  </div>
</template>

<style scoped>
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in-up {
  animation: fadeInUp 0.6s ease-out forwards;
}
</style>
