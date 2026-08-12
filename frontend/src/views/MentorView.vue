<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import AlertModal from '../components/AlertModal.vue'

const router = useRouter()
const currentUser = ref(JSON.parse(localStorage.getItem('user') || '{}'))

const analytics = ref(null)
const groupsList = ref([])
const loading = ref(true)
const error = ref('')

const searchQuery = ref('')
const selectedGroupFilter = ref('ALL')
const sortBy = ref('xp_desc')

// Modals state
const showCreateGroupModal = ref(false)
const isEditingGroup = ref(false)
const editingGroupId = ref(null)
const groupForm = ref({ school_name: '', class: '', academic_year: '' })

const showManageRosterModal = ref(false)
const activeGroupForRoster = ref(null)
const rosterSearch = ref('')

const showDeleteConfirm = ref(false)
const userToDelete = ref(null)
const groupToDelete = ref(null)
const alertState = ref({ show: false, message: '' })

// Active chart tab ('students' or 'groups')
const activeChartTab = ref('groups')
const selectedChartGroup = ref('ALL')

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
    return
  }

  if (currentUser.value.role !== 'Mentor' && currentUser.value.role !== 'Supervisor' && currentUser.value.role !== 'Admin') {
    router.push('/welcome')
    return
  }

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  await Promise.all([fetchAnalytics(), fetchGroups()])
})

const fetchAnalytics = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/mentor/analytics`)
    analytics.value = res.data
  } catch (err) {
    error.value = 'Failed to load mentor analytics. Please try again.'
  } finally {
    loading.value = false
  }
}

const fetchGroups = async () => {
  try {
    const res = await axios.get(`${API_BASE_URL}/api/groups`)
    groupsList.value = res.data
  } catch (err) {
    console.error('Failed to load groups list:', err)
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

const maxGroupXP = computed(() => {
  if (!analytics.value || !analytics.value.groups || analytics.value.groups.length === 0) return 100
  const max = Math.max(...analytics.value.groups.map(g => g.average_xp))
  return max > 0 ? Math.ceil(max) : 100
})

const maxStudentXP = computed(() => {
  if (!analytics.value || !analytics.value.students) return 100
  let list = analytics.value.students
  if (selectedChartGroup.value !== 'ALL') {
    list = list.filter(s => s.group_name === selectedChartGroup.value)
  }
  const max = Math.max(...list.map(s => s.xp), 100)
  return max > 0 ? max : 100
})

// Data for Group vs Group Comparison Chart (Positive Cartesian Axis Columns)
const groupComparisonData = computed(() => {
  if (!analytics.value || !analytics.value.groups) return []
  const maxXP = maxGroupXP.value || 100
  return analytics.value.groups.map(g => ({
    ...g,
    xpPercent: Math.max(8, Math.min(100, Math.round((g.average_xp / maxXP) * 100)))
  }))
})

// Data for Student Comparison Chart (Positive Cartesian Axis Columns)
const studentChartData = computed(() => {
  if (!analytics.value || !analytics.value.students) return []
  let list = analytics.value.students
  if (selectedChartGroup.value !== 'ALL') {
    list = list.filter(s => s.group_name === selectedChartGroup.value)
  }
  const sorted = [...list].sort((a, b) => b.xp - a.xp).slice(0, 8)
  const maxXP = maxStudentXP.value || 100
  return sorted.map(s => ({
    ...s,
    xpPercent: Math.max(8, Math.min(100, Math.round((s.xp / maxXP) * 100)))
  }))
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

// Group Creation & Editing
const openCreateGroupModal = () => {
  isEditingGroup.value = false
  editingGroupId.value = null
  groupForm.value = { school_name: '', class: '', academic_year: `${new Date().getFullYear()}-${new Date().getFullYear() + 1}` }
  showCreateGroupModal.value = true
}

const openEditGroupModal = (g) => {
  isEditingGroup.value = true
  editingGroupId.value = g.id
  groupForm.value = { school_name: g.school_name || g.name.split(' - ')[0] || '', class: g.class || g.name.split(' - ')[1] || '', academic_year: g.academic_year || '2025-2026' }
  showCreateGroupModal.value = true
}

const handleGroupSubmit = async () => {
  try {
    if (isEditingGroup.value) {
      await axios.put(`${API_BASE_URL}/api/groups/${editingGroupId.value}`, groupForm.value)
    } else {
      await axios.post(`${API_BASE_URL}/api/groups`, groupForm.value)
    }
    showCreateGroupModal.value = false
    await Promise.all([fetchAnalytics(), fetchGroups()])
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to save group' }
  }
}

const confirmDeleteGroup = (g) => {
  groupToDelete.value = g
  showDeleteConfirm.value = true
}

const executeGroupDelete = async () => {
  if (!groupToDelete.value) return
  try {
    await axios.delete(`${API_BASE_URL}/api/groups/${groupToDelete.value.id}`)
    showDeleteConfirm.value = false
    groupToDelete.value = null
    await Promise.all([fetchAnalytics(), fetchGroups()])
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to delete group' }
  }
}

// Roster Management Modal
const openManageRosterModal = (g) => {
  activeGroupForRoster.value = g
  showManageRosterModal.value = true
}

const groupStudents = computed(() => {
  if (!activeGroupForRoster.value || !analytics.value?.students) return []
  return analytics.value.students.filter(s => s.group_name === activeGroupForRoster.value.name)
})

const unassignedStudents = computed(() => {
  if (!analytics.value?.students) return []
  return analytics.value.students.filter(s => s.group_name === 'Unassigned' || !s.group_name)
})

const assignStudent = async (studentId, groupId) => {
  try {
    await axios.post(`${API_BASE_URL}/api/groups/assign-student`, {
      user_id: studentId,
      student_group_id: groupId
    })
    await fetchAnalytics()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to update student assignment' }
  }
}

const confirmDeleteUser = (u) => {
  userToDelete.value = u
  showDeleteConfirm.value = true
}

const executeUserDelete = async () => {
  if (!userToDelete.value) return
  try {
    await axios.delete(`${API_BASE_URL}/api/users/${userToDelete.value.id}`)
    showDeleteConfirm.value = false
    userToDelete.value = null
    await fetchAnalytics()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to delete user account' }
  }
}
</script>

<template>
  <div class="w-full max-w-7xl mx-auto px-4 py-6 space-y-8 text-slate-100 font-sans" dir="ltr">
    
    <!-- Page Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-slate-800/60 backdrop-blur-xl p-6 rounded-3xl border border-slate-700/60 shadow-xl">
      <div class="flex items-center gap-3">
        <div class="p-3 rounded-2xl bg-gradient-to-tr from-indigo-500 to-purple-500 text-white shadow-lg shadow-indigo-500/30">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 012-2h2a2 2 0 012 2v6m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
          </svg>
        </div>
        <div>
          <h1 class="text-3xl font-black tracking-tight text-white">Mentor Command Center</h1>
          <p class="text-slate-400 text-sm mt-0.5">Track student progress, manage assigned groups, and analyze live coding metrics</p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button 
          @click="openCreateGroupModal"
          class="flex items-center gap-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-bold py-2.5 px-5 rounded-2xl shadow-lg shadow-indigo-500/25 transition-all active:scale-95 cursor-pointer"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          <span>Create Group</span>
        </button>

        <button 
          @click="fetchAnalytics" 
          :disabled="loading"
          class="flex items-center gap-2 bg-slate-700/80 hover:bg-slate-600/80 text-white font-bold py-2.5 px-4 rounded-2xl border border-slate-600/60 shadow-md transition-all active:scale-95 disabled:opacity-50 cursor-pointer"
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
      <p class="text-slate-400 font-medium">Loading mentor dashboard metrics...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="p-6 bg-rose-500/10 border border-rose-500/30 rounded-2xl text-center text-rose-400">
      {{ error }}
    </div>

    <template v-else-if="analytics">
      
      <!-- Key Metric Cards (STUDENTS ONLY) -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <!-- Total Students -->
        <div class="relative overflow-hidden bg-slate-800/60 backdrop-blur-md p-5 rounded-3xl border border-slate-700/60 shadow-xl group hover:border-indigo-500/50 transition-all duration-300">
          <div class="flex justify-between items-start">
            <div>
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Total Students</p>
              <h3 class="text-3xl font-black text-white mt-1">{{ analytics.total_students }}</h3>
            </div>
            <div class="p-3 bg-indigo-500/10 text-indigo-400 rounded-2xl border border-indigo-500/20 group-hover:scale-110 transition-transform">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
              </svg>
            </div>
          </div>
          <p class="text-xs text-slate-500 mt-3 font-medium">Excludes mentors & admins</p>
        </div>

        <!-- Active Students (7 Days) -->
        <div class="relative overflow-hidden bg-slate-800/60 backdrop-blur-md p-5 rounded-3xl border border-slate-700/60 shadow-xl group hover:border-emerald-500/50 transition-all duration-300">
          <div class="flex justify-between items-start">
            <div>
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Active This Week</p>
              <h3 class="text-3xl font-black text-emerald-400 mt-1">{{ analytics.active_students }}</h3>
            </div>
            <div class="p-3 bg-emerald-500/10 text-emerald-400 rounded-2xl border border-emerald-500/20 group-hover:scale-110 transition-transform">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
              </svg>
            </div>
          </div>
          <p class="text-xs text-slate-500 mt-3 font-medium">Coding within last 7 days</p>
        </div>

        <!-- My Managed Groups -->
        <div class="relative overflow-hidden bg-slate-800/60 backdrop-blur-md p-5 rounded-3xl border border-slate-700/60 shadow-xl group hover:border-purple-500/50 transition-all duration-300">
          <div class="flex justify-between items-start">
            <div>
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">My Groups</p>
              <h3 class="text-3xl font-black text-purple-400 mt-1">{{ analytics.groups.length }}</h3>
            </div>
            <div class="p-3 bg-purple-500/10 text-purple-400 rounded-2xl border border-purple-500/20 group-hover:scale-110 transition-transform">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
              </svg>
            </div>
          </div>
          <p class="text-xs text-slate-500 mt-3 font-medium">Assigned classrooms</p>
        </div>

        <!-- Overall Pass Rate -->
        <div class="relative overflow-hidden bg-slate-800/60 backdrop-blur-md p-5 rounded-3xl border border-slate-700/60 shadow-xl group hover:border-amber-500/50 transition-all duration-300">
          <div class="flex justify-between items-start">
            <div>
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Overall Pass Rate</p>
              <h3 class="text-3xl font-black text-amber-400 mt-1">{{ analytics.overall_pass_rate }}%</h3>
            </div>
            <div class="p-3 bg-amber-500/10 text-amber-400 rounded-2xl border border-amber-500/20 group-hover:scale-110 transition-transform">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
          </div>
          <p class="text-xs text-slate-500 mt-3 font-medium">Across {{ analytics.total_submissions }} submissions</p>
        </div>
      </div>

      <!-- Groups Management Cards Grid -->
      <div class="bg-slate-800/60 backdrop-blur-md p-6 rounded-3xl border border-slate-700/60 shadow-xl space-y-4">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
          <div>
            <h2 class="text-xl font-bold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
              My Managed Groups
            </h2>
            <p class="text-slate-400 text-xs mt-0.5">Classes and student cohorts under your supervision</p>
          </div>
          <button 
            @click="openCreateGroupModal"
            class="text-xs font-semibold px-4 py-2 bg-indigo-600/80 hover:bg-indigo-600 text-white rounded-xl transition-all cursor-pointer"
          >
            + Add New Group
          </button>
        </div>

        <div v-if="analytics.groups.length === 0" class="p-8 text-center bg-slate-900/40 rounded-2xl border border-slate-700/40 text-slate-400">
          No groups found. Click "Add New Group" above to create your first class!
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div 
            v-for="g in analytics.groups" 
            :key="g.id"
            class="bg-slate-900/60 p-5 rounded-2xl border border-slate-700/60 shadow-md space-y-4 hover:border-indigo-500/40 transition-all group"
          >
            <div class="flex justify-between items-start">
              <div>
                <h3 class="font-bold text-white text-base group-hover:text-indigo-400 transition-colors">{{ g.name }}</h3>
                <span class="inline-block text-[11px] font-medium text-slate-400 bg-slate-800 px-2.5 py-0.5 rounded-md mt-1 border border-slate-700">
                  {{ g.student_count }} Students
                </span>
              </div>
              <div class="flex items-center gap-1">
                <button 
                  @click="openEditGroupModal(g)" 
                  class="p-1.5 text-slate-400 hover:text-indigo-400 hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
                  title="Edit Group"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>
                </button>
                <button 
                  @click="confirmDeleteGroup(g)" 
                  class="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
                  title="Delete Group"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                </button>
              </div>
            </div>

            <!-- Group Stats Preview -->
            <div class="grid grid-cols-2 gap-2 text-xs bg-slate-800/50 p-3 rounded-xl border border-slate-700/40">
              <div>
                <p class="text-slate-400">Avg XP</p>
                <p class="text-sm font-black text-amber-400 mt-0.5">{{ g.average_xp }} XP</p>
              </div>
              <div>
                <p class="text-slate-400">Questions Solved</p>
                <p class="text-sm font-black text-emerald-400 mt-0.5">{{ g.passed_count }}</p>
              </div>
            </div>

            <!-- Roster Action -->
            <button 
              @click="openManageRosterModal(g)"
              class="w-full flex items-center justify-center gap-2 py-2 px-3 bg-slate-800 hover:bg-slate-700 text-indigo-300 font-semibold text-xs rounded-xl border border-slate-700 transition-colors cursor-pointer"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
              <span>Manage Student Roster</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Performance Comparison Charts Section -->
      <div class="bg-slate-800/60 backdrop-blur-md p-6 rounded-3xl border border-slate-700/60 shadow-xl space-y-6">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 border-b border-slate-700/60 pb-4">
          <div>
            <h2 class="text-xl font-bold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 012-2h2a2 2 0 012 2v6m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
              Performance Comparison Analytics
            </h2>
            <p class="text-slate-400 text-xs mt-0.5">Visual charts comparing group averages and individual student progress</p>
          </div>

          <!-- Chart Selector Tabs -->
          <div class="flex items-center bg-slate-900/80 p-1 rounded-2xl border border-slate-700/60 text-xs font-semibold">
            <button 
              @click="activeChartTab = 'groups'" 
              :class="activeChartTab === 'groups' ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-400 hover:text-white'"
              class="px-4 py-2 rounded-xl transition-all cursor-pointer"
            >
              Group vs Group
            </button>
            <button 
              @click="activeChartTab = 'students'" 
              :class="activeChartTab === 'students' ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-400 hover:text-white'"
              class="px-4 py-2 rounded-xl transition-all cursor-pointer"
            >
              Student Rankings
            </button>
          </div>
        </div>

        <!-- Chart View A: Group vs Group Comparison (Positive Cartesian Axes Columns) -->
        <div v-if="activeChartTab === 'groups'" class="space-y-4">
          <div class="flex justify-between items-center text-xs">
            <p class="text-slate-400">Vertical Column Chart comparing Average XP per group (Y-Axis = XP, X-Axis = Groups):</p>
            <span class="font-bold text-amber-400 text-[11px]">Max: {{ maxGroupXP }} XP</span>
          </div>
          
          <div v-if="groupComparisonData.length === 0" class="p-8 text-center text-slate-500 text-xs">
            No groups available to compare.
          </div>

          <div v-else class="relative w-full bg-slate-900/90 rounded-2xl p-5 border border-slate-700/60 shadow-inner">
            <!-- Cartesian Grid & Chart Area -->
            <div class="flex items-stretch h-64 gap-3">
              <!-- Y-Axis (المحور الصادي +) -->
              <div class="flex flex-col justify-between items-end text-[10px] font-mono text-slate-400 pr-2 border-r-2 border-slate-600/80 pb-7 select-none">
                <span>{{ maxGroupXP }}</span>
                <span>{{ Math.round(maxGroupXP * 0.75) }}</span>
                <span>{{ Math.round(maxGroupXP * 0.5) }}</span>
                <span>{{ Math.round(maxGroupXP * 0.25) }}</span>
                <span>0</span>
              </div>

              <!-- Content Area with Grid lines and Columns -->
              <div class="relative flex-1 flex flex-col justify-between">
                <!-- Background Horizontal Grid Lines -->
                <div class="absolute inset-0 flex flex-col justify-between pointer-events-none pb-7">
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <!-- X-Axis Baseline (المحور السيني +) -->
                  <div class="border-b-2 border-slate-600/80 w-full"></div>
                </div>

                <!-- Vertical Column Bars -->
                <div class="relative z-10 flex-1 flex items-end justify-around px-2 pb-7 gap-2">
                  <div v-for="g in groupComparisonData" :key="g.id" class="flex-1 flex flex-col items-center h-full justify-end group">
                    <!-- Value tooltip badge over column -->
                    <div class="mb-1 transition-transform group-hover:scale-105">
                      <span class="text-[10px] font-black text-amber-400 bg-slate-950 px-2 py-0.5 rounded-md border border-amber-500/30 shadow-md">
                        {{ g.average_xp }} XP
                      </span>
                    </div>

                    <!-- Column Bar extending vertically upwards -->
                    <div class="w-full max-w-[56px] h-full flex items-end bg-slate-950/40 rounded-t-xl overflow-hidden p-0.5 border border-slate-700/40 group-hover:border-indigo-400/80 transition-all">
                      <div 
                        class="w-full bg-gradient-to-t from-indigo-600 via-purple-500 to-amber-400 rounded-t-lg transition-all duration-700 relative"
                        :style="{ height: g.xpPercent + '%' }"
                      >
                        <div class="absolute inset-0 bg-white/20 opacity-0 group-hover:opacity-100 rounded-t-lg transition-opacity"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- X-Axis Labels -->
                <div class="flex justify-around items-center pt-2 px-2 gap-2 text-xs font-bold text-slate-300">
                  <div v-for="g in groupComparisonData" :key="g.id" class="flex-1 text-center truncate" :title="g.name">
                    {{ g.name }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Chart View B: Student Performance Rankings (Positive Cartesian Axes Columns) -->
        <div v-else class="space-y-4">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
            <p class="text-xs text-slate-400">Vertical Column Chart for Top Performing Students (Y-Axis = XP, X-Axis = Students):</p>
            <div class="flex items-center gap-3">
              <span class="font-bold text-amber-400 text-[11px]">Max: {{ maxStudentXP }} XP</span>
              <select 
                v-model="selectedChartGroup" 
                class="bg-slate-900 border border-slate-700 text-xs rounded-xl px-3 py-1.5 text-slate-200 outline-none focus:border-indigo-500"
              >
                <option value="ALL">All Groups</option>
                <option v-for="g in analytics.groups" :key="g.id" :value="g.name">{{ g.name }}</option>
              </select>
            </div>
          </div>

          <div v-if="studentChartData.length === 0" class="p-8 text-center text-slate-500 text-xs">
            No student data found for this selection.
          </div>

          <div v-else class="relative w-full bg-slate-900/90 rounded-2xl p-5 border border-slate-700/60 shadow-inner">
            <!-- Cartesian Grid & Chart Area -->
            <div class="flex items-stretch h-64 gap-3">
              <!-- Y-Axis (المحور الصادي +) -->
              <div class="flex flex-col justify-between items-end text-[10px] font-mono text-slate-400 pr-2 border-r-2 border-slate-600/80 pb-7 select-none">
                <span>{{ maxStudentXP }}</span>
                <span>{{ Math.round(maxStudentXP * 0.75) }}</span>
                <span>{{ Math.round(maxStudentXP * 0.5) }}</span>
                <span>{{ Math.round(maxStudentXP * 0.25) }}</span>
                <span>0</span>
              </div>

              <!-- Content Area with Grid lines and Columns -->
              <div class="relative flex-1 flex flex-col justify-between">
                <!-- Background Horizontal Grid Lines -->
                <div class="absolute inset-0 flex flex-col justify-between pointer-events-none pb-7">
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <div class="border-b border-slate-700/40 w-full"></div>
                  <!-- X-Axis Baseline (المحور السيني +) -->
                  <div class="border-b-2 border-slate-600/80 w-full"></div>
                </div>

                <!-- Vertical Column Bars -->
                <div class="relative z-10 flex-1 flex items-end justify-around px-2 pb-7 gap-2">
                  <div v-for="st in studentChartData" :key="st.id" class="flex-1 flex flex-col items-center h-full justify-end group">
                    <!-- Value badge over column -->
                    <div class="mb-1 transition-transform group-hover:scale-105">
                      <span class="text-[10px] font-black text-amber-400 bg-slate-950 px-2 py-0.5 rounded-md border border-amber-500/30 shadow-md">
                        {{ st.xp }} XP
                      </span>
                    </div>

                    <!-- Column Bar extending vertically upwards -->
                    <div class="w-full max-w-[48px] h-full flex items-end bg-slate-950/40 rounded-t-xl overflow-hidden p-0.5 border border-slate-700/40 group-hover:border-emerald-400/80 transition-all">
                      <div 
                        class="w-full bg-gradient-to-t from-emerald-600 via-indigo-500 to-purple-400 rounded-t-lg transition-all duration-700 relative"
                        :style="{ height: st.xpPercent + '%' }"
                      >
                        <div class="absolute inset-0 bg-white/20 opacity-0 group-hover:opacity-100 rounded-t-lg transition-opacity"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- X-Axis Labels -->
                <div class="flex justify-around items-center pt-2 px-2 gap-2 text-xs font-bold text-slate-300">
                  <div v-for="st in studentChartData" :key="st.id" class="flex-1 text-center truncate" :title="st.full_name">
                    {{ st.full_name.split(' ')[0] }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Student Roster Table -->
      <div class="bg-slate-800/60 backdrop-blur-md p-6 rounded-3xl border border-slate-700/60 shadow-xl space-y-6">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
          <div>
            <h2 class="text-xl font-bold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
              </svg>
              Student Performance Roster
            </h2>
            <p class="text-slate-400 text-xs mt-0.5">Filter, search, and manage registered students</p>
          </div>

          <!-- Controls -->
          <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
            <!-- Search -->
            <div class="relative flex-1 md:w-64">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="Search student name or email..."
                class="w-full bg-slate-900/80 border border-slate-700/80 rounded-2xl py-2 px-4 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
              />
            </div>

            <!-- Group Filter -->
            <select 
              v-model="selectedGroupFilter"
              class="bg-slate-900/80 border border-slate-700/80 rounded-2xl py-2 px-3 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            >
              <option value="ALL">All Groups</option>
              <option v-for="g in analytics.groups" :key="g.id" :value="g.name">{{ g.name }}</option>
              <option value="Unassigned">Unassigned</option>
            </select>

            <!-- Sort By -->
            <select 
              v-model="sortBy"
              class="bg-slate-900/80 border border-slate-700/80 rounded-2xl py-2 px-3 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            >
              <option value="xp_desc">Sort by Highest XP</option>
              <option value="xp_asc">Sort by Lowest XP</option>
              <option value="passed_desc">Most Solved Questions</option>
              <option value="rate_desc">Highest Success Rate</option>
            </select>
          </div>
        </div>

        <!-- Student Table -->
        <div class="overflow-x-auto rounded-2xl border border-slate-700/60">
          <table class="w-full text-left text-xs text-slate-300">
            <thead class="bg-slate-900/80 text-slate-400 font-semibold uppercase tracking-wider border-b border-slate-700/60">
              <tr>
                <th class="py-3.5 px-4">Student</th>
                <th class="py-3.5 px-4">Group</th>
                <th class="py-3.5 px-4">Level & XP</th>
                <th class="py-3.5 px-4">Solved</th>
                <th class="py-3.5 px-4">Success Rate</th>
                <th class="py-3.5 px-4">Last Active</th>
                <th class="py-3.5 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-700/40 bg-slate-900/30">
              <tr v-if="filteredStudents.length === 0">
                <td colspan="7" class="py-8 text-center text-slate-500">
                  No students found matching the selected filters.
                </td>
              </tr>

              <tr v-for="s in filteredStudents" :key="s.id" class="hover:bg-slate-800/40 transition-colors">
                <!-- Name & Email -->
                <td class="py-3.5 px-4">
                  <div class="flex items-center gap-3">
                    <img 
                      :src="s.avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + s.username" 
                      class="w-9 h-9 rounded-xl bg-slate-800 border border-slate-700" 
                      alt="Avatar"
                    />
                    <div>
                      <p class="font-bold text-white text-sm">{{ s.full_name }}</p>
                      <p class="text-[11px] text-slate-400 font-mono">{{ s.username }} • {{ s.email }}</p>
                    </div>
                  </div>
                </td>

                <!-- Group Badge -->
                <td class="py-3.5 px-4">
                  <span 
                    class="inline-block px-2.5 py-1 rounded-lg text-[11px] font-semibold border"
                    :class="s.group_name === 'Unassigned' ? 'bg-slate-800 text-slate-400 border-slate-700' : 'bg-indigo-500/10 text-indigo-300 border-indigo-500/30'"
                  >
                    {{ s.group_name }}
                  </span>
                </td>

                <!-- Level & XP -->
                <td class="py-3.5 px-4">
                  <div>
                    <span class="font-black text-amber-400">{{ s.xp }} XP</span>
                    <p class="text-[11px] text-slate-400">Level {{ s.level }} ({{ s.level_title }})</p>
                  </div>
                </td>

                <!-- Solved -->
                <td class="py-3.5 px-4">
                  <span class="font-bold text-emerald-400">{{ s.passed_questions }} questions</span>
                </td>

                <!-- Success Rate -->
                <td class="py-3.5 px-4">
                  <div class="flex items-center gap-2">
                    <div class="w-12 bg-slate-800 rounded-full h-1.5 overflow-hidden">
                      <div class="bg-indigo-500 h-full rounded-full" :style="{ width: s.success_rate + '%' }"></div>
                    </div>
                    <span class="font-mono font-bold">{{ s.success_rate }}%</span>
                  </div>
                </td>

                <!-- Last Active -->
                <td class="py-3.5 px-4 font-mono text-slate-400 text-[11px]">
                  {{ formatRelativeTime(s.last_active) }}
                </td>

                <!-- Actions -->
                <td class="py-3.5 px-4 text-right">
                  <button 
                    v-if="currentUser.role === 'Admin'"
                    @click="confirmDeleteUser(s)"
                    class="text-rose-400 hover:text-rose-300 p-1.5 hover:bg-rose-500/10 rounded-lg transition-colors cursor-pointer"
                    title="Delete Account"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Live Code Activity Feed -->
      <div class="bg-slate-800/60 backdrop-blur-md p-6 rounded-3xl border border-slate-700/60 shadow-xl space-y-4">
        <h2 class="text-xl font-bold text-white flex items-center gap-2">
          <svg class="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
          </svg>
          Live Activity Feed
        </h2>

        <div v-if="analytics.activity_feed.length === 0" class="p-6 text-center text-slate-500 text-xs">
          No recent code submissions detected.
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <div 
            v-for="act in analytics.activity_feed" 
            :key="act.id" 
            class="flex items-center gap-3 p-3.5 bg-slate-900/60 rounded-2xl border border-slate-700/40"
          >
            <img 
              :src="act.student_avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + act.student_id" 
              class="w-8 h-8 rounded-lg bg-slate-800 border border-slate-700" 
              alt="Avatar"
            />
            <div class="flex-1 min-w-0 text-xs">
              <p class="font-bold text-white truncate">{{ act.student_name }}</p>
              <p class="text-slate-400 truncate">{{ act.target_title }}</p>
            </div>
            <span 
              class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase"
              :class="act.status === 'Passed' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-rose-500/20 text-rose-300 border border-rose-500/30'"
            >
              {{ act.status }}
            </span>
          </div>
        </div>
      </div>

    </template>

    <!-- Create / Edit Group Modal -->
    <div v-if="showCreateGroupModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
      <div class="bg-slate-900 border border-slate-700/80 rounded-3xl p-6 w-full max-w-md space-y-6 shadow-2xl animate-scale-up">
        <div class="flex justify-between items-center border-b border-slate-800 pb-4">
          <h3 class="text-lg font-bold text-white">{{ isEditingGroup ? 'Edit Group' : 'Create New Group' }}</h3>
          <button @click="showCreateGroupModal = false" class="text-slate-400 hover:text-white cursor-pointer">&times;</button>
        </div>

        <form @submit.prevent="handleGroupSubmit" class="space-y-4 text-xs">
          <div>
            <label class="block text-slate-300 font-semibold mb-1">School / Organization Name</label>
            <input 
              v-model="groupForm.school_name" 
              type="text" 
              required 
              placeholder="e.g. Greenwood High School"
              class="w-full bg-slate-800 border border-slate-700 rounded-xl p-3 text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Class / Grade Name</label>
            <input 
              v-model="groupForm.class" 
              type="text" 
              required 
              placeholder="e.g. Grade 10 - Section A"
              class="w-full bg-slate-800 border border-slate-700 rounded-xl p-3 text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Academic Year</label>
            <input 
              v-model="groupForm.academic_year" 
              type="text" 
              required 
              placeholder="e.g. 2025-2026"
              class="w-full bg-slate-800 border border-slate-700 rounded-xl p-3 text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div class="flex justify-end gap-3 pt-4 border-t border-slate-800">
            <button type="button" @click="showCreateGroupModal = false" class="px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl font-semibold cursor-pointer">Cancel</button>
            <button type="submit" class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold cursor-pointer shadow-lg shadow-indigo-500/25">Save Group</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Manage Student Roster Modal -->
    <div v-if="showManageRosterModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
      <div class="bg-slate-900 border border-slate-700/80 rounded-3xl p-6 w-full max-w-2xl space-y-6 shadow-2xl animate-scale-up">
        <div class="flex justify-between items-center border-b border-slate-800 pb-4">
          <div>
            <h3 class="text-lg font-bold text-white">Group Roster: {{ activeGroupForRoster?.name }}</h3>
            <p class="text-xs text-slate-400">Add unassigned students or remove current group members</p>
          </div>
          <button @click="showManageRosterModal = false" class="text-slate-400 hover:text-white cursor-pointer">&times;</button>
        </div>

        <div class="space-y-6 max-h-[60vh] overflow-y-auto pr-1">
          <!-- Current Group Members -->
          <div class="space-y-2">
            <h4 class="text-xs font-bold uppercase tracking-wider text-indigo-400">Current Members ({{ groupStudents.length }})</h4>
            <div v-if="groupStudents.length === 0" class="p-4 bg-slate-800/40 rounded-xl text-center text-xs text-slate-500">
              No students currently assigned to this group.
            </div>
            <div v-else class="space-y-1.5">
              <div v-for="st in groupStudents" :key="st.id" class="flex justify-between items-center p-2.5 bg-slate-800/60 rounded-xl border border-slate-700/50 text-xs">
                <div class="flex items-center gap-2">
                  <img :src="st.avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + st.username" class="w-6 h-6 rounded-md bg-slate-900" />
                  <span class="font-bold text-white">{{ st.full_name }}</span>
                  <span class="text-slate-400 font-mono">({{ st.email }})</span>
                </div>
                <button 
                  @click="assignStudent(st.id, null)" 
                  class="px-2.5 py-1 bg-rose-500/20 hover:bg-rose-500/30 text-rose-300 text-[11px] font-semibold rounded-lg transition-colors cursor-pointer"
                >
                  Remove
                </button>
              </div>
            </div>
          </div>

          <!-- Unassigned Students -->
          <div class="space-y-2">
            <h4 class="text-xs font-bold uppercase tracking-wider text-emerald-400">Available Unassigned Students ({{ unassignedStudents.length }})</h4>
            <div v-if="unassignedStudents.length === 0" class="p-4 bg-slate-800/40 rounded-xl text-center text-xs text-slate-500">
              No unassigned students available.
            </div>
            <div v-else class="space-y-1.5">
              <div v-for="st in unassignedStudents" :key="st.id" class="flex justify-between items-center p-2.5 bg-slate-800/60 rounded-xl border border-slate-700/50 text-xs">
                <div class="flex items-center gap-2">
                  <img :src="st.avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + st.username" class="w-6 h-6 rounded-md bg-slate-900" />
                  <span class="font-bold text-white">{{ st.full_name }}</span>
                  <span class="text-slate-400 font-mono">({{ st.email }})</span>
                </div>
                <button 
                  @click="assignStudent(st.id, activeGroupForRoster.id)" 
                  class="px-2.5 py-1 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-300 text-[11px] font-semibold rounded-lg transition-colors cursor-pointer"
                >
                  + Add to Group
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-4 border-t border-slate-800">
          <button @click="showManageRosterModal = false" class="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl cursor-pointer">Done</button>
        </div>
      </div>
    </div>

    <!-- Confirm Delete Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
      <div class="bg-slate-900 border border-slate-700/80 rounded-3xl p-6 w-full max-w-sm text-center space-y-4 shadow-2xl">
        <h3 class="text-lg font-bold text-white">Confirm Deletion</h3>
        <p class="text-xs text-slate-300">
          Are you sure you want to delete 
          <span class="font-bold text-rose-400">{{ userToDelete ? userToDelete.full_name : groupToDelete?.name }}</span>?
        </p>
        <div class="flex justify-center gap-3 pt-2">
          <button @click="showDeleteConfirm = false; userToDelete = null; groupToDelete = null;" class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs rounded-xl font-semibold cursor-pointer">Cancel</button>
          <button @click="userToDelete ? executeUserDelete() : executeGroupDelete()" class="px-5 py-2 bg-rose-600 hover:bg-rose-500 text-white text-xs rounded-xl font-bold cursor-pointer">Delete</button>
        </div>
      </div>
    </div>

    <AlertModal 
      :show="alertState.show" 
      :message="alertState.message" 
      @close="alertState.show = false" 
    />
  </div>
</template>

<style scoped>
@keyframes scaleUp {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
.animate-scale-up {
  animation: scaleUp 0.2s ease-out forwards;
}
</style>
