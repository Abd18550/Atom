<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const router = useRouter()
const groups = ref([])
const loading = ref(true)
const error = ref('')

const showModal = ref(false)
const form = ref({
  school_name: '',
  class: '',
  academic_year: ''
})

const submitError = ref('')
const submitSuccess = ref('')

onMounted(async () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  
  if (!token || !userStr) {
    router.push('/login')
    return
  }

  const user = JSON.parse(userStr)
  if (user.role === 'Student') {
    router.push('/welcome')
    return
  }

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  await fetchGroups()
})

const fetchGroups = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/groups`)
    groups.value = res.data
  } catch (err) {
    error.value = 'Failed to load groups.'
  } finally {
    loading.value = false
  }
}

const openModal = () => {
  form.value = {
    school_name: '',
    class: '',
    academic_year: ''
  }
  submitError.value = ''
  submitSuccess.value = ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const submitForm = async () => {
  submitError.value = ''
  submitSuccess.value = ''
  try {
    const res = await axios.post(`${API_BASE_URL}/api/groups`, form.value)
    submitSuccess.value = `Group created successfully!`
    setTimeout(() => {
      closeModal()
      router.push(`/groups/${res.data.group_id}`)
    }, 1000)
  } catch (err) {
    if (err.response && err.response.data && err.response.data.error) {
       submitError.value = err.response.data.error
    } else {
       submitError.value = 'Failed to create group.'
    }
  }
}

const goToGroup = (id) => {
  router.push(`/groups/${id}`)
}
</script>

<template>
  <div class="w-full max-w-6xl">
    <div class="mb-6 flex justify-between items-center transition-colors">
      <div>
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Student Groups</h2>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage classes and academic years</p>
      </div>
      <div class="flex space-x-3">
        <button @click="openModal" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-indigo-600 to-violet-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
          <svg class="-ml-1 mr-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          Create Group
        </button>
      </div>
    </div>

    <!-- Groups Table -->
    <div class="bg-white dark:bg-slate-800 shadow-md overflow-hidden sm:rounded-2xl border border-slate-200 dark:border-slate-700 transition-colors duration-300">
      <div v-if="loading" class="p-8 text-center text-slate-500 dark:text-slate-400">Loading groups...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500 dark:text-red-400">{{ error }}</div>
      <div v-else-if="groups.length === 0" class="p-8 text-center text-slate-500 dark:text-slate-400">No groups found. Create one to get started!</div>
      <table v-else class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
        <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700 transition-colors duration-300">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">School Name</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Class</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Academic Year</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-100 dark:divide-slate-700/50">
          <tr v-for="g in groups" :key="g.id" @click="goToGroup(g.id)" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors cursor-pointer group">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-gray-900 dark:text-slate-200">{{ g.school_name }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-gray-900 dark:text-slate-300">{{ g.class }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300">
                {{ g.academic_year }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Group Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal" aria-hidden="true"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg w-full">
          <form @submit.prevent="submitForm">
            <div class="bg-white dark:bg-slate-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4 transition-colors">
              <div class="sm:flex sm:items-start">
                <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-indigo-100 dark:bg-indigo-900/50">
                  <svg class="h-6 w-6 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
                </div>
                <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
                  <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white" id="modal-title">
                    Create New Group
                  </h3>
                  
                  <div class="mt-4 space-y-4">
                    <div v-if="submitError" class="text-sm text-red-600 bg-red-50 p-2 rounded">{{ submitError }}</div>
                    <div v-if="submitSuccess" class="text-sm text-green-600 bg-green-50 p-2 rounded">{{ submitSuccess }}</div>

                    <div>
                      <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">School Name</label>
                      <input v-model="form.school_name" type="text" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>
                    
                    <div class="grid grid-cols-2 gap-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Class</label>
                        <input v-model="form.class" type="text" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                      </div>
                      <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Academic Year</label>
                        <input v-model="form.academic_year" type="text" placeholder="e.g. 2023-2024" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="bg-gray-50 dark:bg-slate-900/50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse border-t border-gray-200 dark:border-slate-700 transition-colors">
              <button type="submit" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
                Create
              </button>
              <button type="button" @click="closeModal" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 dark:border-slate-600 shadow-sm px-4 py-2 bg-white dark:bg-slate-800 text-base font-medium text-gray-700 dark:text-slate-300 hover:bg-gray-50 dark:hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm transition-colors">
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

  </div>
</template>
