<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import AlertModal from '../components/AlertModal.vue'

const router = useRouter()
const analytics = ref(null)
const loading = ref(true)
const error = ref('')

const activeTab = ref('roster') // 'roster', 'group_comparison', 'manage_groups'

const searchQuery = ref('')
const selectedGroupFilter = ref('ALL')
const sortBy = ref('xp_desc')

const currentUser = ref(null)

// Group Modal State
const showGroupModal = ref(false)
const isEditingGroup = ref(false)
const currentGroupId = ref(null)
const groupForm = ref({
  school_name: '',
  class: '',
  academic_year: ''
})

// Student Detail Modal
const showStudentModal = ref(false)
const selectedStudent = ref(null)

// Delete Confirm Modal State
const showDeleteConfirm = ref(false)
const itemToDelete = ref(null) // { type: 'student'|'group', data: object }
const alertState = ref({ show: false, message: '' })

onMounted(async () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  
  if (!token || !userStr) {
    router.push('/login')
    return
  }

  currentUser.value = JSON.parse(userStr)
  if (currentUser.value.role !== 'Mentor' && currentUser.value.role !== 'Supervisor' && currentUser.value.role !== 'Admin') {
    router.push('/welcome')
    return
  }

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  await fetchAnalytics()
})

const fetchAnalytics = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/mentor/analytics`)
    analytics.value = res.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load mentor analytics. Please try again.'
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
    
    const matchesGroup = selectedGroupFilter.value === 'ALL' || s.group_name === selectedGroupFilter.value
    
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

// Active group statistics for selected group in comparison view
const selectedGroupObj = computed(() => {
  if (!analytics.value || !analytics.value.groups) return null
  if (selectedGroupFilter.value === 'ALL') return null
  return analytics.value.groups.find(g => g.name === selectedGroupFilter.value) || null
})

// Students in currently selected group for comparison chart
const studentsInSelectedGroup = computed(() => {
  if (!analytics.value || !analytics.value.students) return []
  if (selectedGroupFilter.value === 'ALL') return analytics.value.students
  return analytics.value.students.filter(s => s.group_name === selectedGroupFilter.value)
})

// Max XP across groups for responsive chart scaling
const maxGroupXP = computed(() => {
  if (!analytics.value || !analytics.value.groups || analytics.value.groups.length === 0) return 100
  const max = Math.max(...analytics.value.groups.map(g => g.average_xp))
  return max > 0 ? max : 100
})

const maxStudentXPInGroup = computed(() => {
  if (studentsInSelectedGroup.value.length === 0) return 100
  const max = Math.max(...studentsInSelectedGroup.value.map(s => s.xp))
  return max > 0 ? max : 100
})

const formatRelativeTime = (dateStr) => {
  if (!dateStr) return 'Never'
  const date = new Date(dateStr)
  const now = new Date()
  const diffInSeconds = Math.floor((now - date) / 1000)
  
  if (diffInSeconds < 60) return 'Just now'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`
  return `${Math.floor(diffInSeconds / 86400)}d ago`
}

// Group Modal Handlers
const openCreateGroupModal = () => {
  isEditingGroup.value = false
  currentGroupId.value = null
  groupForm.value = { school_name: '', class: '', academic_year: '' }
  showGroupModal.value = true
}

const openEditGroupModal = (group) => {
  isEditingGroup.value = true
  currentGroupId.value = group.id
  groupForm.value = {
    school_name: group.school_name,
    class: group.class,
    academic_year: group.academic_year
  }
  showGroupModal.value = true
}

const saveGroup = async () => {
  if (!groupForm.value.school_name.trim() || !groupForm.value.class.trim() || !groupForm.value.academic_year.trim()) return

  try {
    if (isEditingGroup.value) {
      await axios.put(`${API_BASE_URL}/api/groups/${currentGroupId.value}`, groupForm.value)
    } else {
      await axios.post(`${API_BASE_URL}/api/groups`, groupForm.value)
    }
    showGroupModal.value = false
    await fetchAnalytics()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to save group' }
  }
}

// Student Detail View
const openStudentDetail = (student) => {
  selectedStudent.value = student
  showStudentModal.value = true
}

// Deletion Handlers
const confirmDeleteStudent = (student) => {
  itemToDelete.value = { type: 'student', data: student }
  showDeleteConfirm.value = true
}

const confirmDeleteGroup = (group) => {
  itemToDelete.value = { type: 'group', data: group }
  showDeleteConfirm.value = true
}

const cancelDelete = () => {
  showDeleteConfirm.value = false
  itemToDelete.value = null
}

const executeDelete = async () => {
  if (!itemToDelete.value) return
  
  try {
    if (itemToDelete.value.type === 'student') {
      await axios.delete(`${API_BASE_URL}/api/users/${itemToDelete.value.data.id}`)
      if (showStudentModal.value && selectedStudent.value?.id === itemToDelete.value.data.id) {
        showStudentModal.value = false
      }
    } else if (itemToDelete.value.type === 'group') {
      await axios.delete(`${API_BASE_URL}/api/groups/${itemToDelete.value.data.id}`)
    }
    await fetchAnalytics()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to delete item' }
  } finally {
    cancelDelete()
  }
}

const canManageGroup = (group) => {
  if (!currentUser.value) return false
  if (currentUser.value.role === 'Admin' || currentUser.value.role === 'Supervisor') return true
  return group.created_by_user_id === currentUser.value.id
}
</script>

<template>
  <div class="w-full max-w-7xl mx-auto px-4 py-6 space-y-8 text-slate-100 font-sans" dir="ltr">
    
    <!-- Page Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-slate-800/60 backdrop-blur-xl p-6 rounded-3xl border border-slate-700/60 shadow-xl">
      <div>
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-gradient-to-tr from-indigo-500 to-purple-500 text-white shadow-lg shadow-indigo-500/30">
            <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 012-2h2a2 2 0 012 2v6m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
          </div>
          <div>
            <h1 class="text-3xl font-black tracking-tight text-white">Mentor Analytics & Group Control</h1>
            <p class="text-slate-400 text-sm mt-0.5">Track student metrics, group performance, and manage assigned cohorts</p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button 
          @click="openCreateGroupModal"
          class="flex items-center gap-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-bold py-2.5 px-5 rounded-2xl shadow-lg shadow-indigo-600/30 transition-all active:scale-95 cursor-pointer text-sm"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"></path></svg>
          <span>Create New Group</span>
        </button>

        <button 
          @click="fetchAnalytics" 
          :disabled="loading"
          class="flex items-center gap-2 bg-slate-700/80 hover:bg-slate-600/80 text-white font-bold py-2.5 px-4 rounded-2xl border border-slate-600/60 shadow-md transition-all active:scale-95 disabled:opacity-50 cursor-pointer text-sm"
        >
          <svg :class="{ 'animate-spin': loading }" class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
          </svg>
          <span>Refresh</span>
        </button>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex border-b border-slate-700/60 gap-4">
      <button 
        @click="activeTab = 'roster'"
        :class="['pb-3 px-4 font-bold text-sm transition-colors border-b-2 cursor-pointer', activeTab === 'roster' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-400 hover:text-slate-200']"
      >
        Student Roster & Activity
      </button>
      <button 
        @click="activeTab = 'group_comparison'"
        :class="['pb-3 px-4 font-bold text-sm transition-colors border-b-2 cursor-pointer', activeTab === 'group_comparison' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-400 hover:text-slate-200']"
      >
        Group Comparison & Charts
      </button>
      <button 
        @click="activeTab = 'manage_groups'"
        :class="['pb-3 px-4 font-bold text-sm transition-colors border-b-2 cursor-pointer', activeTab === 'manage_groups' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-400 hover:text-slate-200']"
      >
        Group Management
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading && !analytics" class="p-16 text-center">
      <div class="inline-block animate-spin rounded-full h-12 w-12 border-4 border-indigo-500 border-t-transparent mb-4"></div>
      <p class="text-slate-400 font-medium">Loading mentor dashboard...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-500/10 border border-red-500/30 p-6 rounded-2xl text-center text-red-400">
      {{ error }}
    </div>

    <!-- Dashboard Content -->
    <template v-else-if="analytics">
      
      <!-- Top KPI Cards (STRICTLY STUDENT-ONLY STATS) -->
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
          <div class="text-xs text-slate-400 mt-2 font-medium">Students enrolled (excluding mentors)</div>
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
            {{ analytics.total_students > 0 ? Math.round((analytics.active_students / analytics.total_students) * 100) : 0 }}% Weekly engagement
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
          <div class="text-xs text-slate-400 mt-2 font-medium">Executions by students</div>
        </div>

        <!-- Overall Pass Rate Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg relative overflow-hidden group hover:border-teal-500/40 transition-all">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-teal-500/10 rounded-full blur-2xl group-hover:bg-teal-500/20 transition-all"></div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-slate-400 text-sm font-semibold">Accuracy Rate</span>
            <div class="p-2.5 rounded-2xl bg-teal-500/20 text-teal-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </div>
          </div>
          <div class="text-4xl font-black text-white tracking-tight">{{ analytics.overall_pass_rate }}%</div>
          <div class="text-xs text-teal-400 mt-2 font-medium">Average solution accuracy</div>
        </div>

      </div>

      <!-- TAB 1: STUDENT ROSTER & ACTIVITY -->
      <div v-if="activeTab === 'roster'" class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        <!-- Left 2 Columns: Student Roster Table -->
        <div class="lg:col-span-2 space-y-5">
          
          <!-- Controls Bar -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-5 rounded-3xl shadow-lg flex flex-col sm:flex-row gap-4 justify-between items-center">
            
            <!-- Search Input -->
            <div class="relative w-full sm:w-72">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="Search student name or email..."
                class="w-full bg-slate-900/90 border border-slate-700/80 text-white rounded-2xl pl-10 pr-4 py-2.5 text-sm focus:outline-none focus:border-indigo-500 transition-colors"
              />
              <svg class="w-5 h-5 text-slate-400 absolute left-3 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
            </div>

            <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
              <!-- Group Filter Dropdown -->
              <select 
                v-model="selectedGroupFilter"
                class="bg-slate-900/90 border border-slate-700/80 text-slate-200 rounded-2xl px-4 py-2.5 text-sm focus:outline-none focus:border-indigo-500 font-medium"
              >
                <option value="ALL">All Groups</option>
                <option v-for="g in analytics.groups" :key="g.id" :value="g.name">{{ g.name }}</option>
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

          <!-- Students Table Card -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 rounded-3xl shadow-xl overflow-hidden">
            <div class="px-6 py-4 border-b border-slate-700/60 flex items-center justify-between">
              <h3 class="text-lg font-bold text-white flex items-center gap-2">
                <span>Student Performance Roster</span>
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
                    <th class="px-6 py-4">Accuracy</th>
                    <th class="px-6 py-4">Last Active</th>
                    <th class="px-6 py-4 text-center">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-700/40 text-sm">
                  <tr v-for="s in filteredStudents" :key="s.id" class="hover:bg-slate-700/30 transition-colors">
                    
                    <!-- Student Info -->
                    <td class="px-6 py-4 whitespace-nowrap">
                      <div class="flex items-center gap-3 cursor-pointer" @click="openStudentDetail(s)">
                        <div class="w-10 h-10 rounded-2xl bg-indigo-900/40 border border-indigo-500/30 flex items-center justify-center font-bold text-indigo-400 overflow-hidden shrink-0">
                          <img v-if="s.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + s.avatar" class="w-full h-full object-cover bg-slate-900" />
                          <span v-else>{{ s.full_name ? s.full_name.charAt(0) : 'S' }}</span>
                        </div>
                        <div>
                          <div class="font-bold text-white hover:text-indigo-400 transition-colors">{{ s.full_name }}</div>
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
                      <div class="flex items-center justify-center gap-1">
                        <button 
                          @click="openStudentDetail(s)"
                          class="p-2 text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/10 rounded-xl transition-all cursor-pointer"
                          title="View Performance Analytics"
                        >
                          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>
                        </button>
                        <button 
                          @click="confirmDeleteStudent(s)" 
                          class="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition-all cursor-pointer"
                          title="Delete Student Account"
                        >
                          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                        </button>
                      </div>
                    </td>

                  </tr>
                </tbody>
              </table>

              <div v-if="filteredStudents.length === 0" class="p-8 text-center text-slate-400 text-sm">
                No students match the selected filter.
              </div>
            </div>

          </div>

        </div>

        <!-- Right 1 Column: Group Overview & Activity Stream -->
        <div class="space-y-6">
          
          <!-- Groups Overview Card -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-bold text-white">Your Student Groups</h3>
              <span class="text-xs text-slate-400 font-normal">Total: {{ analytics.groups.length }}</span>
            </div>

            <div class="space-y-3">
              <div 
                v-for="g in analytics.groups" 
                :key="g.id"
                @click="selectedGroupFilter = (selectedGroupFilter === g.name ? 'ALL' : g.name)"
                :class="{
                  'p-4 rounded-2xl border transition-all cursor-pointer': true,
                  'bg-indigo-500/10 border-indigo-500/50 shadow-md': selectedGroupFilter === g.name,
                  'bg-slate-900/60 border-slate-700/60 hover:border-slate-600': selectedGroupFilter !== g.name
                }"
              >
                <div class="flex items-center justify-between mb-2">
                  <div class="font-bold text-white text-sm">{{ g.name }}</div>
                  <span class="text-xs bg-slate-700/80 text-slate-300 font-bold px-2 py-0.5 rounded-lg">
                    {{ g.student_count }} students
                  </span>
                </div>

                <div class="flex items-center justify-between text-xs text-slate-400">
                  <span>Avg XP: <b class="text-amber-400 font-mono">{{ g.average_xp }} XP</b></span>
                  <span>Tasks Solved: <b class="text-emerald-400 font-mono">{{ g.total_passed_questions }}</b></span>
                </div>
              </div>

              <div v-if="analytics.groups.length === 0" class="text-xs text-slate-500 text-center py-4">
                No groups created yet. Click "Create New Group" to get started.
              </div>
            </div>
          </div>

          <!-- Live Activity Stream Widget -->
          <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl">
            <h3 class="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping"></span>
              <span>Live Submission Activity</span>
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
                No recent activity recorded.
              </div>
            </div>
          </div>

        </div>

      </div>

      <!-- TAB 2: GROUP COMPARISON & CUSTOM CHARTS -->
      <div v-else-if="activeTab === 'group_comparison'" class="space-y-8">
        
        <!-- Group Comparison Bar Chart Card -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
            <div>
              <h3 class="text-xl font-bold text-white">Comparative Group Performance Chart</h3>
              <p class="text-xs text-slate-400 mt-1">Comparing average XP and total solved questions across your student cohorts</p>
            </div>
          </div>

          <div v-if="analytics.groups.length === 0" class="p-12 text-center text-slate-400">
            No groups available to compare.
          </div>

          <!-- Group Visual Chart Grid -->
          <div v-else class="space-y-6">
            <div v-for="g in analytics.groups" :key="g.id" class="space-y-2">
              <div class="flex justify-between items-center text-sm">
                <div class="font-bold text-slate-200 flex items-center gap-2">
                  <span>{{ g.name }}</span>
                  <span class="text-xs text-slate-400 font-normal">({{ g.student_count }} students)</span>
                </div>
                <div class="font-mono text-xs text-amber-400 font-bold">
                  Avg XP: {{ g.average_xp }} | Total Tasks: {{ g.total_passed_questions }}
                </div>
              </div>

              <!-- Bar 1: Average XP -->
              <div class="w-full bg-slate-900/80 rounded-xl h-6 p-1 flex items-center">
                <div 
                  class="bg-gradient-to-r from-indigo-500 to-purple-500 h-full rounded-lg transition-all duration-500 flex items-center justify-end pr-2 text-[10px] font-black text-white"
                  :style="{ width: Math.max((g.average_xp / maxGroupXP) * 100, 8) + '%' }"
                >
                  {{ g.average_xp }} XP
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Student Comparison Within Selected Group -->
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl space-y-6">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
            <div>
              <h3 class="text-xl font-bold text-white">Student Leaderboard & Comparison within Group</h3>
              <p class="text-xs text-slate-400 mt-1">Select a group to analyze individual student performance relative to group mates</p>
            </div>

            <!-- Group Selector -->
            <select 
              v-model="selectedGroupFilter"
              class="bg-slate-900/90 border border-slate-700/80 text-slate-200 rounded-2xl px-4 py-2 text-sm focus:outline-none focus:border-indigo-500 font-medium"
            >
              <option value="ALL">All Students</option>
              <option v-for="g in analytics.groups" :key="g.id" :value="g.name">{{ g.name }}</option>
            </select>
          </div>

          <!-- Students Bar Chart Comparison -->
          <div class="space-y-4 pt-2">
            <div v-for="s in studentsInSelectedGroup" :key="s.id" class="p-4 bg-slate-900/50 rounded-2xl border border-slate-700/50 space-y-2">
              <div class="flex justify-between items-center">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-xl bg-indigo-900/40 border border-indigo-500/30 flex items-center justify-center font-bold text-indigo-400 overflow-hidden shrink-0">
                    <img v-if="s.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + s.avatar" class="w-full h-full object-cover" />
                    <span v-else>{{ s.full_name ? s.full_name.charAt(0) : 'S' }}</span>
                  </div>
                  <div>
                    <span class="font-bold text-slate-200 text-sm">{{ s.full_name }}</span>
                    <span class="text-xs text-slate-400 ml-2">({{ s.group_name }})</span>
                  </div>
                </div>
                <div class="font-mono text-xs text-emerald-400 font-bold">
                  {{ s.xp }} XP | {{ s.passed_questions }} Solved
                </div>
              </div>

              <!-- Bar Comparison -->
              <div class="w-full bg-slate-950 rounded-lg h-5 p-0.5 flex items-center">
                <div 
                  class="bg-gradient-to-r from-emerald-500 to-teal-400 h-full rounded-md transition-all duration-500 flex items-center justify-end pr-2 text-[10px] font-black text-slate-950"
                  :style="{ width: Math.max((s.xp / maxStudentXPInGroup) * 100, 10) + '%' }"
                >
                  {{ s.xp }} XP
                </div>
              </div>
            </div>

            <div v-if="studentsInSelectedGroup.length === 0" class="p-8 text-center text-slate-400 text-sm">
              No students found for this group selection.
            </div>
          </div>
        </div>

      </div>

      <!-- TAB 3: GROUP MANAGEMENT -->
      <div v-else-if="activeTab === 'manage_groups'" class="space-y-6">
        <div class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-xl flex justify-between items-center">
          <div>
            <h3 class="text-xl font-bold text-white">Manage Cohort Groups</h3>
            <p class="text-xs text-slate-400 mt-1">Create and configure groups. You have administrative permissions on groups created by you.</p>
          </div>
          <button 
            @click="openCreateGroupModal"
            class="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold py-2 px-4 rounded-xl text-sm transition-all"
          >
            + Create Group
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div 
            v-for="g in analytics.groups" 
            :key="g.id" 
            class="bg-slate-800/80 backdrop-blur border border-slate-700/70 p-6 rounded-3xl shadow-lg flex flex-col justify-between space-y-4"
          >
            <div>
              <div class="flex justify-between items-start mb-2">
                <h4 class="text-lg font-bold text-white">{{ g.school_name }}</h4>
                <span :class="['px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider', canManageGroup(g) ? 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/30' : 'bg-slate-700 text-slate-400']">
                  {{ canManageGroup(g) ? 'Owner' : 'View Only' }}
                </span>
              </div>
              <div class="text-sm font-semibold text-indigo-400 mb-1">Class: {{ g.class }}</div>
              <div class="text-xs text-slate-400">Academic Year: {{ g.academic_year }}</div>
              <div class="text-xs text-slate-400 mt-1">Created by: <b class="text-slate-300">{{ g.created_by_name }}</b></div>
            </div>

            <div class="pt-4 border-t border-slate-700/50 flex justify-between items-center">
              <span class="text-xs font-mono font-bold text-slate-300">{{ g.student_count }} Enrolled Students</span>
              
              <div v-if="canManageGroup(g)" class="flex gap-2">
                <button 
                  @click="openEditGroupModal(g)" 
                  class="px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded-lg text-xs font-semibold transition-all"
                >
                  Edit
                </button>
                <button 
                  @click="confirmDeleteGroup(g)" 
                  class="px-3 py-1.5 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-lg text-xs font-semibold transition-all"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>

          <div v-if="analytics.groups.length === 0" class="col-span-full p-12 text-center text-slate-400">
            No student groups registered yet.
          </div>
        </div>
      </div>

    </template>

    <!-- CREATE / EDIT GROUP MODAL -->
    <div v-if="showGroupModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen p-4">
        <div class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm transition-opacity" @click="showGroupModal = false"></div>
        
        <div class="relative bg-slate-800 rounded-3xl max-w-md w-full p-6 border border-slate-700 shadow-2xl space-y-5">
          <div class="flex justify-between items-center border-b border-slate-700/60 pb-3">
            <h3 class="text-lg font-bold text-white">{{ isEditingGroup ? 'Edit Group' : 'Create New Group' }}</h3>
            <button @click="showGroupModal = false" class="text-slate-400 hover:text-white">&times;</button>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-1">School Name</label>
              <input 
                v-model="groupForm.school_name" 
                type="text" 
                placeholder="e.g. Al-Amal High School"
                class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-1">Class / Grade</label>
              <input 
                v-model="groupForm.class" 
                type="text" 
                placeholder="e.g. Grade 10 - Section A"
                class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-1">Academic Year</label>
              <input 
                v-model="groupForm.academic_year" 
                type="text" 
                placeholder="e.g. 2025/2026"
                class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div class="flex justify-end gap-3 pt-3">
            <button @click="showGroupModal = false" class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded-xl text-sm font-semibold">
              Cancel
            </button>
            <button @click="saveGroup" class="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-sm font-bold shadow-lg">
              Save Group
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- STUDENT DETAIL PERFORMANCE COMPARISON MODAL -->
    <div v-if="showStudentModal && selectedStudent" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen p-4">
        <div class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm transition-opacity" @click="showStudentModal = false"></div>
        
        <div class="relative bg-slate-800 rounded-3xl max-w-xl w-full p-6 border border-slate-700 shadow-2xl space-y-6">
          
          <!-- Header -->
          <div class="flex items-center justify-between border-b border-slate-700/60 pb-4">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-2xl bg-indigo-900/40 border border-indigo-500/30 flex items-center justify-center font-bold text-indigo-400 overflow-hidden">
                <img v-if="selectedStudent.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + selectedStudent.avatar" class="w-full h-full object-cover" />
                <span v-else>{{ selectedStudent.full_name.charAt(0) }}</span>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">{{ selectedStudent.full_name }}</h3>
                <div class="text-xs text-slate-400">@{{ selectedStudent.username }} | {{ selectedStudent.group_name }}</div>
              </div>
            </div>
            <button @click="showStudentModal = false" class="text-slate-400 hover:text-white text-xl font-bold">&times;</button>
          </div>

          <!-- Comparative Performance Chart (Requirement 7) -->
          <div class="space-y-4">
            <h4 class="text-sm font-bold text-indigo-300 uppercase tracking-wider">Performance vs Group Average</h4>

            <!-- XP Comparison -->
            <div class="space-y-1">
              <div class="flex justify-between text-xs text-slate-300">
                <span>XP Points</span>
                <span class="font-mono">Student: <b>{{ selectedStudent.xp }}</b> | Group Avg: <b>{{ selectedStudent.group_avg_xp }}</b></span>
              </div>
              <div class="w-full bg-slate-900 rounded-lg h-4 flex overflow-hidden">
                <div class="bg-indigo-500 h-full transition-all" :style="{ width: Math.min((selectedStudent.xp / Math.max(selectedStudent.xp, selectedStudent.group_avg_xp, 1)) * 100, 100) + '%' }"></div>
              </div>
            </div>

            <!-- Tasks Solved Comparison -->
            <div class="space-y-1">
              <div class="flex justify-between text-xs text-slate-300">
                <span>Tasks Solved</span>
                <span class="font-mono">Student: <b>{{ selectedStudent.passed_questions }}</b> | Group Avg: <b>{{ selectedStudent.group_avg_passed }}</b></span>
              </div>
              <div class="w-full bg-slate-900 rounded-lg h-4 flex overflow-hidden">
                <div class="bg-emerald-500 h-full transition-all" :style="{ width: Math.min((selectedStudent.passed_questions / Math.max(selectedStudent.passed_questions, selectedStudent.group_avg_passed, 1)) * 100, 100) + '%' }"></div>
              </div>
            </div>

            <!-- Accuracy Rate -->
            <div class="space-y-1">
              <div class="flex justify-between text-xs text-slate-300">
                <span>Submission Accuracy</span>
                <span class="font-mono"><b>{{ selectedStudent.success_rate }}%</b></span>
              </div>
              <div class="w-full bg-slate-900 rounded-lg h-4 flex overflow-hidden">
                <div class="bg-teal-400 h-full transition-all" :style="{ width: selectedStudent.success_rate + '%' }"></div>
              </div>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex justify-between items-center pt-4 border-t border-slate-700/60">
            <button @click="confirmDeleteStudent(selectedStudent)" class="px-4 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-xl text-xs font-semibold">
              Delete Account
            </button>
            <button @click="showStudentModal = false" class="px-5 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-xl text-xs font-bold">
              Close
            </button>
          </div>

        </div>
      </div>
    </div>

    <!-- DELETE CONFIRMATION MODAL -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen p-4">
        <div class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm transition-opacity" @click="cancelDelete"></div>
        <div class="relative bg-slate-800 rounded-3xl max-w-md w-full p-6 border border-slate-700 shadow-2xl text-left" dir="ltr">
          <div class="flex items-start gap-4">
            <div class="p-3 bg-red-500/20 rounded-2xl text-red-400">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white mb-1">Confirm Deletion</h3>
              <p class="text-sm text-slate-400">
                Are you sure you want to delete <b class="text-white">{{ itemToDelete?.data?.full_name || itemToDelete?.data?.name || itemToDelete?.data?.school_name }}</b>? This action cannot be undone.
              </p>
            </div>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <button @click="cancelDelete" class="px-5 py-2.5 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded-xl text-sm font-semibold transition-all cursor-pointer">
              Cancel
            </button>
            <button @click="executeDelete" class="px-5 py-2.5 bg-red-600 hover:bg-red-500 text-white rounded-xl text-sm font-bold shadow-lg shadow-red-600/30 transition-all cursor-pointer">
              Confirm Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <AlertModal :show="alertState.show" :message="alertState.message" @close="alertState.show = false" />
  </div>
</template>
