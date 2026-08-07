<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import AlertModal from '../components/AlertModal.vue'

const router = useRouter()
const analytics = ref(null)
const groupComparisons = ref([])
const loading = ref(true)
const error = ref('')

const searchQuery = ref('')
const selectedGroupFilter = ref('ALL') // 'ALL' or group ID as number/string
const sortBy = ref('xp_desc')

const showDeleteConfirm = ref(false)
const userToDelete = ref(null)
const alertState = ref({ show: false, message: '' })

onMounted(async () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  
  if (!token || !userStr) {
    router.push('/login')
    return
  }

  const user = JSON.parse(userStr)
  if (user.role !== 'Mentor' && user.role !== 'Supervisor' && user.role !== 'Admin') {
    router.push('/welcome')
    return
  }

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  await fetchAllData()
})

const fetchAllData = async () => {
  loading.value = true
  error.value = ''
  try {
    const [analyticsRes, comparisonRes] = await Promise.all([
      axios.get(`${API_BASE_URL}/api/mentor/analytics`),
      axios.get(`${API_BASE_URL}/api/mentor/group-comparison`)
    ])
    analytics.value = analyticsRes.data
    groupComparisons.value = comparisonRes.data
  } catch (err) {
    console.error('Mentor Analytics fetch error:', err)
    const backendErr = err.response?.data?.error || err.response?.data?.message || err.message
    error.value = backendErr ? `Analytics error: ${backendErr}` : 'Failed to load mentor analytics. Please try again.'
  } finally {
    loading.value = false
  }
}

// Filtered and sorted students
const filteredStudents = computed(() => {
  if (!analytics.value || !analytics.value.students) return []
  
  let list = analytics.value.students.filter(s => {
    const matchesSearch = s.full_name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          s.username.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          s.email.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    let matchesGroup = true
    if (selectedGroupFilter.value !== 'ALL') {
      matchesGroup = s.group_id === Number(selectedGroupFilter.value)
    }
    
    return matchesSearch && matchesGroup
  })

  return list.sort((a, b) => {
    if (sortBy.value === 'xp_desc') return b.xp - a.xp
    if (sortBy.value === 'xp_asc') return a.xp - b.xp
    if (sortBy.value === 'passed_desc') return b.passed_questions - a.passed_questions
    if (sortBy.value === 'rate_desc') return b.success_rate - a.success_rate
    return 0
  })
})

// Max values for responsive SVG chart scaling
const maxGroupXP = computed(() => {
  if (!groupComparisons.value || groupComparisons.value.length === 0) return 100
  return Math.max(...groupComparisons.value.map(g => g.average_xp), 100)
})

const maxStudentXPInGroup = computed(() => {
  if (filteredStudents.value.length === 0) return 100
  return Math.max(...filteredStudents.value.map(s => s.xp), 100)
})

const formatRelativeTime = (dateStr) => {
  if (!dateStr) return 'Inactive'
  const date = new Date(dateStr)
  const now = new Date()
  const diffInSeconds = Math.floor((now - date) / 1000)
  
  if (diffInSeconds < 60) return 'Just now'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`
  return `${Math.floor(diffInSeconds / 86400)}d ago`
}

const confirmDelete = (u) => {
  userToDelete.value = u
  showDeleteConfirm.value = true
}

const cancelDelete = () => {
  showDeleteConfirm.value = false
  userToDelete.value = null
}

const executeDelete = async () => {
  if (!userToDelete.value) return
  
  try {
    await axios.delete(`${API_BASE_URL}/api/users/${userToDelete.value.id}`)
    await fetchAllData()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to delete student account' }
  } finally {
    cancelDelete()
  }
}
</script>

<template>
  <div class="w-full max-w-7xl mx-auto px-4 py-6 space-y-8 text-slate-100 font-sans">
    
    <!-- Page Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-slate-800/60 backdrop-blur-xl p-6 rounded-3xl border border-slate-700/60 shadow-xl">
      <div class="flex items-center gap-3">
        <div class="p-2.5 rounded-2xl bg-gradient-to-tr from-indigo-500 to-purple-500 text-white shadow-lg shadow-indigo-500/30">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 012-2h2a2 2 0 012 2v6m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
          </svg>
        </div>
        <div>
          <h1 class="text-3xl font-black tracking-tight text-white">Mentor Analytics Dashboard</h1>
          <p class="text-slate-400 text-sm mt-0.5">Manage your assigned groups, view student metrics & compare performance</p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button 
          @click="router.push('/groups')"
          class="flex items-center gap-2 bg-indigo-600/90 hover:bg-indigo-500 text-white font-bold py-2.5 px-5 rounded-2xl shadow-md transition-all active:scale-95 cursor-pointer"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"></path></svg>
          <span>Manage Groups</span>
        </button>

        <button 
          @click="fetchAllData" 
          :disabled="loading"
          class="flex items-center gap-2 bg-slate-700/80 hover:bg-slate-600/80 text-white font-bold py-2.5 px-5 rounded-2xl border border-slate-600/60 shadow-md transition-all active:scale-95 disabled:opacity-50 cursor-pointer"
        >
          <svg :class="{ 'animate-spin': loading }" class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
          </svg>
          <span>Refresh</span>
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && !analytics" class="p-16 text-center">
      <div class="inline-block animate-spin rounded-full h-12 w-12 border-4 border-indigo-500 border-t-transparent mb-4"></div>
      <p class="text-slate-400 font-medium">Loading mentor analytics...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-500/10 border border-red-500/30 p-6 rounded-2xl text-center text-red-400">
      {{ error }}
    </div>

    <!-- Dashboard Content -->
    <template v-else-if="analytics">
      
      <!-- Top KPI Cards (Students ONLY) -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        
        <!-- Total Students Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg relative overflow-hidden group hover:border-indigo-500/40 transition-all">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-indigo-500/10 rounded-full blur-2xl group-hover:bg-indigo-500/20 transition-all"></div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-slate-400 text-sm font-semibold">Total Students</span>
            <div class="p-2.5 rounded-2xl bg-indigo-500/20 text-indigo-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
            </div>
          </div>
          <div class="text-4xl font-black text-white tracking-tight">{{ analytics.total_students }}</div>
          <div class="text-xs text-slate-400 mt-2 font-medium">Students in your groups</div>
        </div>

        <!-- Active Students Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg relative overflow-hidden group hover:border-emerald-500/40 transition-all">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-emerald-500/10 rounded-full blur-2xl group-hover:bg-emerald-500/20 transition-all"></div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-slate-400 text-sm font-semibold">Active Students (7d)</span>
            <div class="p-2.5 rounded-2xl bg-emerald-500/20 text-emerald-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
            </div>
          </div>
          <div class="text-4xl font-black text-white tracking-tight">{{ analytics.active_students }}</div>
          <div class="text-xs text-emerald-400 mt-2 font-medium">
            {{ analytics.total_students > 0 ? Math.round((analytics.active_students / analytics.total_students) * 100) : 0 }}% active engagement
          </div>
        </div>

        <!-- Total Submissions Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg relative overflow-hidden group hover:border-purple-500/40 transition-all">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-purple-500/10 rounded-full blur-2xl group-hover:bg-purple-500/20 transition-all"></div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-slate-400 text-sm font-semibold">Code Submissions</span>
            <div class="p-2.5 rounded-2xl bg-purple-500/20 text-purple-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg>
            </div>
          </div>
          <div class="text-4xl font-black text-white tracking-tight">{{ analytics.total_submissions }}</div>
          <div class="text-xs text-slate-400 mt-2 font-medium">Executed in sandbox</div>
        </div>

        <!-- Overall Pass Rate Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg relative overflow-hidden group hover:border-teal-500/40 transition-all">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-teal-500/10 rounded-full blur-2xl group-hover:bg-teal-500/20 transition-all"></div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-slate-400 text-sm font-semibold">Overall Pass Rate</span>
            <div class="p-2.5 rounded-2xl bg-teal-500/20 text-teal-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </div>
          </div>
          <div class="text-4xl font-black text-white tracking-tight">{{ analytics.overall_pass_rate }}%</div>
          <div class="text-xs text-teal-400 mt-2 font-medium">Average code accuracy</div>
        </div>

      </div>

      <!-- Group-to-Group Comparison Chart Section -->
      <div v-if="groupComparisons.length > 0" class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-bold text-white">Groups Overview & Comparison</h3>
            <p class="text-xs text-slate-400">Compare average XP and problem solving metrics across your groups</p>
          </div>
          <span class="text-xs bg-indigo-500/20 text-indigo-300 font-bold px-3 py-1 rounded-full border border-indigo-500/30">
            {{ groupComparisons.length }} Groups
          </span>
        </div>

        <div class="space-y-3 pt-2">
          <div v-for="g in groupComparisons" :key="g.id" class="space-y-1.5">
            <div class="flex justify-between items-center text-xs font-semibold">
              <span class="text-slate-200">{{ g.name }} ({{ g.student_count }} students)</span>
              <span class="text-indigo-300 font-mono">{{ g.average_xp }} Avg XP | {{ g.avg_passed }} Avg Solved</span>
            </div>
            <!-- Custom Horizontal SVG Bar -->
            <div class="w-full bg-slate-900/80 rounded-xl h-4 overflow-hidden p-0.5 border border-slate-700/50 flex items-center">
              <div 
                class="bg-gradient-to-r from-indigo-500 to-purple-500 h-full rounded-lg transition-all duration-500"
                :style="{ width: Math.max((g.average_xp / maxGroupXP) * 100, 3) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Section: Controls, Student Comparison Chart & Roster -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        <!-- Left 2 Columns: Student Controls, Comparison Chart & Roster -->
        <div class="lg:col-span-2 space-y-5">
          
          <!-- Controls Bar -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-5 rounded-3xl shadow-lg flex flex-col sm:flex-row gap-4 justify-between items-center">
            
            <!-- Search Input -->
            <div class="relative w-full sm:w-64">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="Search student or email..."
                class="w-full bg-slate-900/90 border border-slate-700/80 text-white rounded-2xl pl-10 pr-4 py-2.5 text-sm focus:outline-none focus:border-indigo-500 transition-colors"
              />
              <svg class="w-5 h-5 text-slate-400 absolute left-3 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
            </div>

            <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
              <!-- Filter by Group Dropdown -->
              <select 
                v-model="selectedGroupFilter"
                class="bg-slate-900/90 border border-slate-700/80 text-slate-200 rounded-2xl px-4 py-2.5 text-sm focus:outline-none focus:border-indigo-500 font-medium"
              >
                <option value="ALL">All Groups</option>
                <option v-for="g in analytics.groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>

              <!-- Sort Dropdown -->
              <select 
                v-model="sortBy"
                class="bg-slate-900/90 border border-slate-700/80 text-slate-200 rounded-2xl px-4 py-2.5 text-sm focus:outline-none focus:border-indigo-500 font-medium"
              >
                <option value="xp_desc">Sort: Highest XP</option>
                <option value="passed_desc">Sort: Questions Solved</option>
                <option value="rate_desc">Sort: Success Rate</option>
              </select>
            </div>

          </div>

          <!-- Student-to-Student Comparison Bar Chart in Selected View -->
          <div v-if="filteredStudents.length > 0" class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl space-y-4">
            <h4 class="text-sm font-bold text-white flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>
              <span>Student Comparison (XP Ranking)</span>
            </h4>
            <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
              <div v-for="s in filteredStudents.slice(0, 8)" :key="s.id" class="space-y-1">
                <div class="flex justify-between text-[11px] font-medium text-slate-300">
                  <span class="truncate max-w-[150px]">{{ s.full_name }}</span>
                  <span class="font-mono text-emerald-400 font-bold">{{ s.xp }} XP</span>
                </div>
                <div class="w-full bg-slate-900 h-2.5 rounded-full overflow-hidden">
                  <div 
                    class="bg-gradient-to-r from-emerald-400 to-teal-500 h-full rounded-full transition-all duration-300"
                    :style="{ width: Math.max((s.xp / maxStudentXPInGroup) * 100, 4) + '%' }"
                  ></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Students Table Card -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 rounded-3xl shadow-xl overflow-hidden">
            <div class="px-6 py-4 border-b border-slate-700/60 flex items-center justify-between">
              <h3 class="text-lg font-bold text-white flex items-center gap-2">
                <span>Student Roster</span>
                <span class="bg-indigo-500/20 text-indigo-400 text-xs font-black px-2.5 py-0.5 rounded-full">
                  {{ filteredStudents.length }}
                </span>
              </h3>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left border-collapse">
                <thead>
                  <tr class="bg-slate-900/50 text-slate-400 text-xs font-bold uppercase tracking-wider border-b border-slate-700/60">
                    <th class="px-6 py-4">Student</th>
                    <th class="px-6 py-4">Group</th>
                    <th class="px-6 py-4">Level & XP</th>
                    <th class="px-6 py-4">Solved</th>
                    <th class="px-6 py-4">Success Rate</th>
                    <th class="px-6 py-4">Last Active</th>
                    <th class="px-6 py-4 text-center">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-700/40 text-sm">
                  <tr v-for="s in filteredStudents" :key="s.id" class="hover:bg-slate-700/30 transition-colors">
                    
                    <!-- Student Info -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <div class="flex items-center gap-3">
                        <div class="w-10 h-10 rounded-2xl bg-indigo-900/40 border border-indigo-500/30 flex items-center justify-center font-bold text-indigo-400 overflow-hidden shrink-0">
                          <img v-if="s.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + s.avatar" class="w-full h-full object-cover bg-slate-900" />
                          <span v-else>{{ s.full_name ? s.full_name.charAt(0) : 'S' }}</span>
                        </div>
                        <div>
                          <div class="font-bold text-white">{{ s.full_name }}</div>
                          <div class="text-xs text-slate-400">@{{ s.username }}</div>
                        </div>
                      </div>
                    </td>

                    <!-- Group Name -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <span class="inline-flex items-center px-3 py-1 rounded-xl text-xs font-bold bg-slate-700/60 text-slate-300 border border-slate-600/50">
                        {{ s.group_name }}
                      </span>
                    </td>

                    <!-- Level & XP -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <div class="flex flex-col">
                        <div class="flex items-center gap-2">
                          <span class="text-xs font-black text-amber-400 bg-amber-400/10 px-2 py-0.5 rounded-md border border-amber-400/20">
                            Lvl {{ s.level }}
                          </span>
                          <span class="text-xs font-semibold text-slate-300">{{ s.level_title }}</span>
                        </div>
                        <div class="text-xs text-indigo-400 font-mono font-bold mt-1">
                          {{ s.xp }} XP
                        </div>
                      </div>
                    </td>

                    <!-- Solved Questions -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <span class="font-mono font-bold text-emerald-400 text-base">
                        {{ s.passed_questions }}
                      </span>
                      <span class="text-slate-500 text-xs ml-1">tasks</span>
                    </td>

                    <!-- Success Rate -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <div class="flex items-center gap-2">
                        <div class="w-16 bg-slate-700 rounded-full h-2 overflow-hidden">
                          <div 
                            class="bg-gradient-to-r from-emerald-500 to-teal-400 h-full rounded-full"
                            :style="{ width: Math.min(s.success_rate, 100) + '%' }"
                          ></div>
                        </div>
                        <span class="font-mono text-xs font-bold text-slate-300">{{ s.success_rate }}%</span>
                      </div>
                    </td>

                    <!-- Last Active -->
                    <td class="px-6 py-4 whitespace-nowrap text-xs text-slate-400 font-medium">
                      {{ formatRelativeTime(s.last_active) }}
                    </td>

                    <!-- Actions -->
                    <td class="px-6 py-4 whitespace-nowrap text-center">
                      <button 
                        @click="confirmDelete(s)" 
                        class="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition-all cursor-pointer"
                        title="Delete Student"
                      >
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                      </button>
                    </td>

                  </tr>
                </tbody>
              </table>

              <div v-if="filteredStudents.length === 0" class="p-8 text-center text-slate-400 text-sm">
                No students found matching your filters.
              </div>
            </div>

          </div>

        </div>

        <!-- Right 1 Column: Group List & Activity Stream -->
        <div class="space-y-6">
          
          <!-- Groups Overview Card -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl">
            <h3 class="text-lg font-bold text-white mb-4 flex items-center justify-between">
              <span>Your Student Groups</span>
              <span class="text-xs text-slate-400 font-normal">Total: {{ analytics.groups.length }}</span>
            </h3>

            <div class="space-y-3">
              <div 
                v-for="g in analytics.groups" 
                :key="g.id"
                @click="selectedGroupFilter = (selectedGroupFilter === g.id ? 'ALL' : g.id)"
                :class="{
                  'p-4 rounded-2xl border transition-all cursor-pointer': true,
                  'bg-indigo-500/10 border-indigo-500/50 shadow-md': selectedGroupFilter === g.id,
                  'bg-slate-900/60 border-slate-700/60 hover:border-slate-600': selectedGroupFilter !== g.id
                }"
              >
                <div class="flex items-center justify-between mb-2">
                  <div class="font-bold text-white text-sm">{{ g.name }}</div>
                  <span class="text-xs bg-slate-700/80 text-slate-300 font-bold px-2 py-0.5 rounded-lg">
                    {{ g.student_count }} students
                  </span>
                </div>

                <div class="flex items-center justify-between text-xs text-slate-400">
                  <span>Average XP: <b class="text-amber-400 font-mono">{{ g.average_xp }} XP</b></span>
                  <span>Passed: <b class="text-emerald-400 font-mono">{{ g.passed_count }}</b></span>
                </div>
              </div>

              <div v-if="analytics.groups.length === 0" class="text-xs text-slate-500 text-center py-4">
                No groups created yet.
              </div>
            </div>
          </div>

          <!-- Live Activity Stream -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl">
            <h3 class="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping"></span>
              <span>Recent Activity Feed</span>
            </h3>

            <div class="space-y-3 max-h-96 overflow-y-auto pr-1">
              <div 
                v-for="act in analytics.activity_feed" 
                :key="act.id"
                class="p-3 bg-slate-900/60 border border-slate-700/50 rounded-2xl flex items-center justify-between gap-3 text-xs"
              >
                <div class="flex items-center gap-3 overflow-hidden">
                  <div class="w-8 h-8 rounded-xl bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-bold text-indigo-400 overflow-hidden shrink-0">
                    <img v-if="act.student_avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + act.student_avatar" class="w-full h-full object-cover" />
                    <span v-else>{{ act.student_name ? act.student_name.charAt(0) : 'S' }}</span>
                  </div>
                  <div class="truncate">
                    <div class="font-bold text-slate-200 truncate">{{ act.student_name }}</div>
                    <div class="text-slate-400 text-[11px] truncate">{{ act.target_title }}</div>
                  </div>
                </div>

                <div class="flex flex-col items-end shrink-0">
                  <span :class="{
                    'px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider': true,
                    'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30': act.status === 'Passed',
                    'bg-red-500/20 text-red-400 border border-red-500/30': act.status === 'Failed' || act.status === 'SyntaxError' || act.status === 'Error',
                    'bg-amber-500/20 text-amber-400 border border-amber-500/30': act.status === 'Pending' || act.status === 'Running'
                  }">
                    {{ act.status }}
                  </span>
                  <span class="text-[10px] text-slate-500 mt-1 font-mono">
                    {{ formatRelativeTime(act.created_at) }}
                  </span>
                </div>
              </div>

              <div v-if="!analytics.activity_feed || analytics.activity_feed.length === 0" class="text-xs text-slate-500 text-center py-4">
                No recent activity.
              </div>
            </div>
          </div>

        </div>

      </div>

    </template>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen p-4">
        <div class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm transition-opacity" @click="cancelDelete"></div>
        <div class="relative bg-slate-800 rounded-3xl max-w-md w-full p-6 border border-slate-700 shadow-2xl text-left">
          <div class="flex items-start gap-4">
            <div class="p-3 bg-red-500/20 rounded-2xl text-red-400 shrink-0">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white mb-1">Delete Student Account</h3>
              <p class="text-sm text-slate-400">
                Are you sure you want to permanently delete <b class="text-white">{{ userToDelete?.full_name }}</b>? This action cannot be undone.
              </p>
            </div>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <button @click="cancelDelete" class="px-5 py-2.5 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded-xl text-sm font-semibold transition-all cursor-pointer">
              Cancel
            </button>
            <button @click="executeDelete" class="px-5 py-2.5 bg-red-600 hover:bg-red-500 text-white rounded-xl text-sm font-bold shadow-lg shadow-red-600/30 transition-all cursor-pointer">
              Delete Account
            </button>
          </div>
        </div>
      </div>
    </div>

    <AlertModal :show="alertState.show" :message="alertState.message" @close="alertState.show = false" />
  </div>
</template>
