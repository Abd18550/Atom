<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const router = useRouter()
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const stats = ref(null)
const groupData = ref(null)
const loading = ref(true)

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
    return
  }

  try {
    const [statsRes, groupRes] = await Promise.all([
      axios.get(`${API_BASE_URL}/api/student-stats`, { headers: { Authorization: `Bearer ${token}` } }),
      axios.get(`${API_BASE_URL}/api/student/group-comparison`, { headers: { Authorization: `Bearer ${token}` } }).catch(() => ({ data: null }))
    ])
    stats.value = statsRes.data
    groupData.value = groupRes?.data || null
  } catch (err) {
    console.error('Failed to load student dashboard data:', err)
  } finally {
    loading.value = false
  }
})

const maxPeerXP = computed(() => {
  if (!groupData.value || !groupData.value.peers || groupData.value.peers.length === 0) return 100
  const max = Math.max(...groupData.value.peers.map(p => p.xp), 100)
  return max > 0 ? max : 100
})

const peerChartData = computed(() => {
  if (!groupData.value || !groupData.value.peers) return []
  const max = maxPeerXP.value || 100
  return groupData.value.peers.map(p => ({
    ...p,
    xpPercent: Math.max(8, Math.min(100, Math.round((p.xp / max) * 100)))
  }))
})

const levelTitles = [
  'Beginner', 'Learner', 'Explorer', 'Coder', 'Developer',
  'Engineer', 'Architect', 'Master', 'Grandmaster', 'Legend'
]

const currentLevelTitle = computed(() => {
  if (!stats.value) return 'Beginner'
  return stats.value.level_title || levelTitles[(stats.value.level || 1) - 1] || 'Beginner'
})

const xpProgress = computed(() => {
  if (!stats.value) return 0
  const currentXP = stats.value.xp || 0
  const currentLevelXP = stats.value.current_level_xp || 0
  const nextLevelXP = stats.value.next_level_xp || 100
  if (nextLevelXP <= currentLevelXP) return 100 // Max level
  return Math.min(100, ((currentXP - currentLevelXP) / (nextLevelXP - currentLevelXP)) * 100)
})

const xpInLevel = computed(() => {
  if (!stats.value) return { current: 0, needed: 100 }
  const currentXP = stats.value.xp || 0
  const currentLevelXP = stats.value.current_level_xp || 0
  const nextLevelXP = stats.value.next_level_xp || 100
  return {
    current: currentXP - currentLevelXP,
    needed: nextLevelXP - currentLevelXP
  }
})

const stageProgress = computed(() => {
  if (!stats.value) return 0
  if (stats.value.total_stages === 0) return 0
  return Math.round((stats.value.completed_stages / stats.value.total_stages) * 100)
})

const greetingMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
})

const motivationalMessage = computed(() => {
  if (!stats.value) return 'Start your journey today!'
  const level = stats.value.level || 1
  if (level >= 8) return '🔥 You are unstoppable! Keep pushing the limits!'
  if (level >= 5) return '🚀 Amazing progress! You are becoming a true developer!'
  if (level >= 3) return '💪 Great work! You are building strong foundations!'
  if (stats.value.passed_questions > 0) return '⚡ Nice start! Keep solving to level up!'
  return '🌟 Begin your coding adventure!'
})

const isNonStudent = computed(() => {
  return user.value.role && user.value.role !== 'Student'
})

const goToDashboard = () => router.push('/welcome')
const goToLearningPath = () => router.push('/learning-path')
const goToExercises = () => router.push('/exercises')
const goToCurrentStage = () => {
  if (stats.value?.current_stage_id) {
    router.push('/stage/' + stats.value.current_stage_id)
  } else {
    router.push('/learning-path')
  }
}
</script>

<template>
  <div class="w-full max-w-5xl space-y-6 animate-fade-in-up">

    <!-- Back to Dashboard (for non-students viewing this page) -->
    <div v-if="isNonStudent" class="flex items-center gap-3 px-1">
      <button
        @click="goToDashboard"
        class="flex items-center gap-2 px-4 py-2.5 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 text-sm font-semibold text-slate-600 dark:text-slate-300 hover:text-indigo-600 dark:hover:text-indigo-400 transition-all duration-200 group"
      >
        <svg class="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
        Back to Dashboard
      </button>
      <span class="text-xs text-slate-400 dark:text-slate-500 font-medium">Viewing as Student</span>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center min-h-[400px]">
      <div class="text-center">
        <div class="inline-block w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin mb-4"></div>
        <p class="text-slate-400 font-medium animate-pulse">Loading your dashboard...</p>
      </div>
    </div>

    <template v-else>
      <!-- Hero Section: Level Banner -->
      <div class="relative bg-gradient-to-br from-indigo-600 via-purple-600 to-fuchsia-600 dark:from-indigo-700 dark:via-purple-700 dark:to-fuchsia-700 rounded-3xl overflow-hidden shadow-2xl shadow-indigo-500/20 dark:shadow-indigo-900/40">
        <!-- Animated Background Effects -->
        <div class="absolute inset-0 overflow-hidden pointer-events-none">
          <div class="absolute top-[-20%] right-[-10%] w-72 h-72 bg-white/10 rounded-full blur-3xl animate-float-slow"></div>
          <div class="absolute bottom-[-15%] left-[-5%] w-56 h-56 bg-white/10 rounded-full blur-3xl animate-float-slow-reverse"></div>
          <div class="absolute top-[40%] left-[50%] w-32 h-32 bg-fuchsia-300/10 rounded-full blur-2xl animate-pulse"></div>
        </div>

        <div class="relative z-10 p-8 md:p-10">
          <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
            <!-- Left: Greeting + Level -->
            <div class="flex-1">
              <p class="text-indigo-200 text-sm font-semibold uppercase tracking-widest mb-1">{{ greetingMessage }}</p>
              <h1 class="text-3xl md:text-4xl font-black text-white tracking-tight mb-2">
                {{ user.full_name || user.username }}
              </h1>
              <p class="text-indigo-100/80 text-sm">{{ motivationalMessage }}</p>
            </div>

            <!-- Right: Level Badge -->
            <div class="flex items-center gap-4">
              <div class="relative">
                <div class="w-20 h-20 md:w-24 md:h-24 rounded-2xl rotate-45 bg-white/15 backdrop-blur-md border border-white/20 flex items-center justify-center shadow-lg">
                  <div class="-rotate-45 text-center">
                    <span class="text-3xl md:text-4xl font-black text-white drop-shadow-lg">{{ stats?.level || 1 }}</span>
                  </div>
                </div>
                <!-- Glow Ring -->
                <div class="absolute inset-0 rounded-2xl rotate-45 border-2 border-white/30 animate-ping opacity-20"></div>
              </div>
              <div>
                <p class="text-white/70 text-xs uppercase tracking-wider font-bold">Level</p>
                <p class="text-white text-lg md:text-xl font-black">{{ currentLevelTitle }}</p>
                <p class="text-indigo-200 text-xs mt-0.5">{{ stats?.xp || 0 }} XP Total</p>
              </div>
            </div>
          </div>

          <!-- XP Progress Bar -->
          <div class="mt-6">
            <div class="flex justify-between text-xs text-white/70 mb-1.5 font-medium">
              <span>Level {{ stats?.level || 1 }} Progress</span>
              <span>{{ xpInLevel.current }} / {{ xpInLevel.needed }} XP</span>
            </div>
            <div class="h-3 bg-white/15 rounded-full overflow-hidden backdrop-blur-sm">
              <div
                class="h-full bg-gradient-to-r from-yellow-300 via-amber-400 to-orange-400 rounded-full transition-all duration-1000 ease-out relative"
                :style="{ width: xpProgress + '%' }"
              >
                <div class="absolute inset-0 bg-gradient-to-r from-transparent to-white/30 rounded-full animate-shimmer"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Stats Grid -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
        <!-- XP Card -->
        <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-lg transition-all duration-300 group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2.5 bg-amber-100 dark:bg-amber-500/15 rounded-xl text-amber-600 dark:text-amber-400 group-hover:scale-110 transition-transform">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
            </div>
          </div>
          <p class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.xp || 0 }}</p>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Total XP</p>
        </div>

        <!-- Level Card -->
        <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-lg transition-all duration-300 group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2.5 bg-purple-100 dark:bg-purple-500/15 rounded-xl text-purple-600 dark:text-purple-400 group-hover:scale-110 transition-transform">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"></path></svg>
            </div>
          </div>
          <p class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.level || 1 }}</p>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Current Level</p>
        </div>

        <!-- Questions Solved -->
        <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-lg transition-all duration-300 group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2.5 bg-emerald-100 dark:bg-emerald-500/15 rounded-xl text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </div>
          </div>
          <p class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.passed_questions || 0 }}</p>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Questions Solved</p>
        </div>

        <!-- Stages Completed -->
        <div class="bg-white dark:bg-slate-800 rounded-2xl p-5 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-lg transition-all duration-300 group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2.5 bg-blue-100 dark:bg-blue-500/15 rounded-xl text-blue-600 dark:text-blue-400 group-hover:scale-110 transition-transform">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
            </div>
          </div>
          <p class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.completed_stages || 0 }}<span class="text-sm font-semibold text-slate-400"> / {{ stats?.total_stages || 0 }}</span></p>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">Stages Completed</p>
        </div>
      </div>

      <!-- Class Group Performance Comparison Chart (2D Cartesian Axes Column Chart) -->
      <div v-if="groupData && groupData.in_group" class="bg-white dark:bg-slate-800 rounded-3xl p-6 border border-slate-100 dark:border-slate-700 shadow-xl space-y-5">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 border-b border-slate-100 dark:border-slate-700/60 pb-4">
          <div>
            <div class="flex items-center gap-2">
              <span class="p-2 bg-indigo-100 dark:bg-indigo-500/15 text-indigo-600 dark:text-indigo-400 rounded-xl">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 012-2h2a2 2 0 012 2v6m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
              </span>
              <h2 class="text-xl font-bold text-slate-900 dark:text-white">Class Performance Chart</h2>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">Comparing your XP against classmates in <span class="font-bold text-indigo-600 dark:text-indigo-400">{{ groupData.group_name }}</span></p>
          </div>

          <!-- Rank & Avg Badge -->
          <div class="flex items-center gap-3">
            <div class="bg-slate-100 dark:bg-slate-700/70 text-slate-700 dark:text-slate-200 px-3 py-1.5 rounded-xl text-xs font-bold">
              Class Avg: {{ groupData.group_avg_xp }} XP
            </div>
            <div class="flex items-center gap-2 bg-gradient-to-r from-amber-500 to-orange-600 text-white px-4 py-1.5 rounded-xl shadow-md text-xs font-black">
              <span>Your Rank: #{{ groupData.my_rank }}</span>
              <span class="text-amber-100 font-normal">of {{ groupData.total_students }}</span>
            </div>
          </div>
        </div>

        <!-- 2D Positive Cartesian Column Chart (Y-Axis = XP, X-Axis = Classmates) -->
        <div class="relative w-full bg-slate-50 dark:bg-slate-900/90 rounded-2xl p-5 border border-slate-200/80 dark:border-slate-700/60 shadow-inner">
          <div class="flex items-stretch h-56 gap-3">
            <!-- Y-Axis (المحور الصادي +) -->
            <div class="flex flex-col justify-between items-end text-[10px] font-mono text-slate-400 dark:text-slate-400 pr-2 border-r-2 border-slate-300 dark:border-slate-600/80 pb-7 select-none">
              <span>{{ maxPeerXP }}</span>
              <span>{{ Math.round(maxPeerXP * 0.75) }}</span>
              <span>{{ Math.round(maxPeerXP * 0.5) }}</span>
              <span>{{ Math.round(maxPeerXP * 0.25) }}</span>
              <span>0</span>
            </div>

            <!-- Chart Columns Container with Grid Lines -->
            <div class="relative flex-1 flex flex-col justify-between">
              <!-- Background Horizontal Grid Lines -->
              <div class="absolute inset-0 flex flex-col justify-between pointer-events-none pb-7">
                <div class="border-b border-slate-200 dark:border-slate-700/40 w-full"></div>
                <div class="border-b border-slate-200 dark:border-slate-700/40 w-full"></div>
                <div class="border-b border-slate-200 dark:border-slate-700/40 w-full"></div>
                <div class="border-b border-slate-200 dark:border-slate-700/40 w-full"></div>
                <!-- X-Axis Baseline (المحور السيني +) -->
                <div class="border-b-2 border-slate-400 dark:border-slate-500/80 w-full"></div>
              </div>

              <!-- Vertical Column Bars -->
              <div class="relative z-10 flex-1 flex items-end justify-around px-2 pb-7 gap-2">
                <div v-for="peer in peerChartData" :key="peer.id" class="flex-1 flex flex-col items-center h-full justify-end group">
                  <!-- Value Tooltip Badge -->
                  <div class="mb-1 transition-transform group-hover:scale-105">
                    <span 
                      class="text-[10px] font-black px-2 py-0.5 rounded-md shadow-sm border"
                      :class="peer.is_me ? 'bg-amber-500 text-white border-amber-400' : 'bg-slate-200 dark:bg-slate-950 text-slate-700 dark:text-indigo-300 border-slate-300 dark:border-slate-700'"
                    >
                      {{ peer.xp }} XP
                    </span>
                  </div>

                  <!-- Column Bar extending vertically upwards -->
                  <div 
                    class="w-full max-w-[48px] h-full flex items-end rounded-t-xl overflow-hidden p-0.5 border transition-all"
                    :class="peer.is_me ? 'bg-amber-500/10 border-amber-400 dark:border-amber-500' : 'bg-slate-200/50 dark:bg-slate-950/40 border-slate-300 dark:border-slate-700/40 group-hover:border-indigo-400'"
                  >
                    <div 
                      class="w-full rounded-t-lg transition-all duration-700 relative"
                      :class="peer.is_me ? 'bg-gradient-to-t from-amber-500 via-orange-500 to-amber-300' : 'bg-gradient-to-t from-indigo-600 via-purple-500 to-indigo-400'"
                      :style="{ height: peer.xpPercent + '%' }"
                    >
                      <div class="absolute inset-0 bg-white/20 opacity-0 group-hover:opacity-100 rounded-t-lg transition-opacity"></div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- X-Axis Labels -->
              <div class="flex justify-around items-center pt-2 px-2 gap-2 text-xs font-bold text-slate-600 dark:text-slate-300">
                <div 
                  v-for="peer in peerChartData" 
                  :key="peer.id" 
                  class="flex-1 text-center truncate flex flex-col items-center" 
                  :title="peer.full_name"
                >
                  <span :class="peer.is_me ? 'text-amber-600 dark:text-amber-400 font-extrabold' : ''">
                    {{ peer.full_name.split(' ')[0] }}
                  </span>
                  <span v-if="peer.is_me" class="text-[9px] bg-amber-500 text-white font-black px-1.5 py-0.2 rounded mt-0.5">YOU</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Action Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

        <!-- Continue Journey / Current Stage Card -->
        <div
          @click="goToCurrentStage"
          class="group relative bg-white dark:bg-slate-800 rounded-2xl p-6 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden"
        >
          <div class="absolute right-0 top-0 w-32 h-32 bg-gradient-to-bl from-rose-100 to-transparent dark:from-rose-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-4 mb-4">
              <div class="p-3.5 bg-gradient-to-br from-rose-500 to-rose-600 dark:from-rose-600 dark:to-rose-700 text-white rounded-xl shadow-md group-hover:shadow-rose-200 dark:group-hover:shadow-rose-900/50 group-hover:scale-110 transition-all">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
              </div>
              <div>
                <h3 class="text-lg font-bold text-slate-800 dark:text-slate-200 group-hover:text-rose-600 dark:group-hover:text-rose-400 transition-colors">Continue Journey</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400" v-if="stats?.current_stage_title">
                  Current: {{ stats.current_stage_title }}
                </p>
                <p class="text-sm text-emerald-500 font-semibold" v-else>
                  ✅ All stages completed!
                </p>
              </div>
            </div>

            <!-- Stage Progress Bar -->
            <div class="mt-2">
              <div class="flex justify-between text-xs text-slate-500 dark:text-slate-400 mb-1">
                <span>Journey Progress</span>
                <span>{{ stageProgress }}%</span>
              </div>
              <div class="h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-rose-400 to-rose-500 rounded-full transition-all duration-700"
                  :style="{ width: stageProgress + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Coding Challenges Card -->
        <div
          @click="goToExercises"
          class="group relative bg-white dark:bg-slate-800 rounded-2xl p-6 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden"
        >
          <div class="absolute right-0 top-0 w-32 h-32 bg-gradient-to-bl from-purple-100 to-transparent dark:from-purple-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-4 mb-4">
              <div class="p-3.5 bg-gradient-to-br from-purple-500 to-purple-600 dark:from-purple-600 dark:to-purple-700 text-white rounded-xl shadow-md group-hover:shadow-purple-200 dark:group-hover:shadow-purple-900/50 group-hover:scale-110 transition-all">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg>
              </div>
              <div>
                <h3 class="text-lg font-bold text-slate-800 dark:text-slate-200 group-hover:text-purple-600 dark:group-hover:text-purple-400 transition-colors">Coding Challenges</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400">Earn +30 XP per challenge</p>
              </div>
            </div>
            <div class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 mt-2">
              <svg class="w-4 h-4 text-emerald-500" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>
              <span>{{ stats?.passed_exercises || 0 }} challenges solved</span>
            </div>
          </div>
        </div>

        <!-- Orbital Map Card -->
        <div
          @click="goToLearningPath"
          class="group relative bg-white dark:bg-slate-800 rounded-2xl p-6 border border-slate-100 dark:border-slate-700 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer overflow-hidden md:col-span-2"
        >
          <div class="absolute right-0 top-0 w-32 h-32 bg-gradient-to-bl from-indigo-100 to-transparent dark:from-indigo-500/10 rounded-bl-full z-0 transition-transform group-hover:scale-125"></div>
          <div class="relative z-10 flex items-center gap-4">
            <div class="p-3.5 bg-gradient-to-br from-indigo-500 to-indigo-600 dark:from-indigo-600 dark:to-indigo-700 text-white rounded-xl shadow-md group-hover:shadow-indigo-200 dark:group-hover:shadow-indigo-900/50 group-hover:scale-110 transition-all">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">Orbital Map</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">View the full learning path and track your orbit</p>
            </div>
            <div class="ml-auto text-slate-400 group-hover:text-indigo-500 group-hover:translate-x-1 transition-all">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
            </div>
          </div>
        </div>
      </div>
    </template>
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

@keyframes float-slow {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-15px); }
}

.animate-float-slow {
  animation: float-slow 6s ease-in-out infinite;
}

.animate-float-slow-reverse {
  animation: float-slow 8s ease-in-out infinite reverse;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(300%); }
}

.animate-shimmer {
  animation: shimmer 2s ease-in-out infinite;
}
</style>
