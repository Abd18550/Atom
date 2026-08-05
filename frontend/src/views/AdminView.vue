<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import AlertModal from '../components/AlertModal.vue'

const router = useRouter()
const users = ref([])
const loading = ref(true)
const error = ref('')

const currentUser = ref(JSON.parse(localStorage.getItem('user') || '{}'))

const showModal = ref(false)
const showDeleteConfirm = ref(false)
const userToDelete = ref(null)

const newUserType = ref('Mentor')
const form = ref({
  username: '',
  email: '',
  password: '',
  role: 'Mentor', // Default 
  full_name: '',
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
  if (user.role !== 'Admin' && user.role !== 'Supervisor') {
    router.push('/welcome')
    return
  }

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  await fetchUsers()
})

const fetchUsers = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/users?exclude_students=true`)
    users.value = res.data
  } catch (err) {
    error.value = 'Failed to load users.'
  } finally {
    loading.value = false
  }
}

const openModal = (type) => {
  newUserType.value = type
  form.value = {
    username: '',
    email: '',
    password: '',
    role: type,
    full_name: '',
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
    const payload = { ...form.value }

    await axios.post(`${API_BASE_URL}/api/users`, payload)
    submitSuccess.value = `${payload.role} created successfully!`
    setTimeout(() => {
      closeModal()
      fetchUsers()
    }, 1500)
  } catch (err) {
    if (err.response && err.response.data && err.response.data.error) {
       submitError.value = err.response.data.error
    } else {
       submitError.value = 'Failed to create user.'
    }
  }
}

const canDelete = (u) => {
  if (u.role === 'Admin') return false; 
  if (currentUser.value.role === 'Admin') return true;
  if (currentUser.value.role === 'Supervisor') {
    return u.role !== 'Supervisor';
  }
  return false;
}

const confirmDelete = (u) => {
  userToDelete.value = u
  showDeleteConfirm.value = true
}

const cancelDelete = () => {
  showDeleteConfirm.value = false
  userToDelete.value = null
}

const alertState = ref({ show: false, message: '' })

const executeDelete = async () => {
  if (!userToDelete.value) return;
  
  try {
    await axios.delete(`${API_BASE_URL}/api/users/${userToDelete.value.id}`)
    fetchUsers()
  } catch (err) {
    alertState.value = { show: true, message: err.response?.data?.error || 'Failed to delete user' };
  } finally {
    cancelDelete()
  }
}
</script>

<template>
  <div class="w-full max-w-6xl">
    <div class="mb-6 flex justify-between items-center transition-colors">
      <div>
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Admin Dashboard</h2>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage platform accounts</p>
      </div>
      <div class="flex space-x-3">
        <button v-if="currentUser.role === 'Admin'" @click="openModal('Supervisor')" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-purple-600 to-fuchsia-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500">
          <svg class="-ml-1 mr-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          Add Supervisor
        </button>
        <button @click="openModal('Mentor')" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-indigo-600 to-violet-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
          <svg class="-ml-1 mr-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          Add Mentor
        </button>
      </div>
    </div>

    <!-- Users Table -->
    <div class="bg-white dark:bg-slate-800 shadow-md overflow-hidden sm:rounded-2xl border border-slate-200 dark:border-slate-700 transition-colors duration-300">
      <div v-if="loading" class="p-8 text-center text-slate-500 dark:text-slate-400">Loading users...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500 dark:text-red-400">{{ error }}</div>
      <table v-else class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
        <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700 transition-colors duration-300">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Name / Username</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Email</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Role</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Created At</th>
            <th scope="col" class="relative px-6 py-3">
              <span class="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-100 dark:divide-slate-700/50">
          <tr v-for="u in users" :key="u.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div class="flex-shrink-0 h-10 w-10 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-700 font-bold uppercase overflow-hidden">
                  <img v-if="u.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + u.avatar" class="w-full h-full object-cover bg-indigo-50" />
                  <template v-else>{{ u.full_name ? u.full_name.charAt(0) : 'U' }}</template>
                </div>
                <div class="ml-4">
                  <div class="text-sm font-medium text-gray-900 dark:text-slate-200">{{ u.full_name }}</div>
                  <div class="text-sm text-gray-500 dark:text-slate-400">@{{ u.username }}</div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-gray-900 dark:text-slate-300">{{ u.email }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                :class="{
                  'bg-red-100 text-red-800': u.role === 'Admin',
                  'bg-purple-600 text-white': u.role === 'Supervisor',
                  'bg-blue-100 text-blue-800': u.role === 'Mentor',
                  'bg-green-100 text-green-800': u.role === 'Student'
                }">
                {{ u.role }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-slate-400">
              {{ new Date(u.created_at).toLocaleDateString() }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button v-if="canDelete(u)" @click="confirmDelete(u)" class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 ml-4 font-semibold p-2">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create User Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal" aria-hidden="true"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <!-- Modal panel -->
        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg w-full">
          <form @submit.prevent="submitForm">
            <div class="bg-white dark:bg-slate-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4 transition-colors">
              <div class="sm:flex sm:items-start">
                <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full"
                     :class="newUserType === 'Mentor' ? 'bg-indigo-100 dark:bg-indigo-900/50' : 'bg-purple-100 dark:bg-purple-900/50'">
                  <svg v-if="newUserType === 'Mentor'" class="h-6 w-6 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                  <svg v-else class="h-6 w-6 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                </div>
                <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
                  <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white" id="modal-title">
                    Create New {{ newUserType }}
                  </h3>
                  
                  <div class="mt-4 space-y-4">
                    <div v-if="submitError" class="text-sm text-red-600 bg-red-50 p-2 rounded">{{ submitError }}</div>
                    <div v-if="submitSuccess" class="text-sm text-green-600 bg-green-50 p-2 rounded">{{ submitSuccess }}</div>

                    <div>
                      <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Full Name</label>
                      <input v-model="form.full_name" type="text" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>
                    
                    <div>
                      <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Username</label>
                      <input v-model="form.username" type="text" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>

                    <div>
                      <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Email</label>
                      <input v-model="form.email" type="email" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>

                    <div>
                      <label class="block text-sm font-medium text-gray-700 dark:text-slate-300">Password</label>
                      <input v-model="form.password" type="password" required class="bg-transparent dark:text-white mt-1 block w-full border border-gray-300 dark:border-slate-600 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
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

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-900 bg-opacity-50 backdrop-blur-sm transition-opacity" @click="cancelDelete" aria-hidden="true"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <!-- Modal panel -->
        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-md w-full border border-slate-100 dark:border-slate-700 p-6 animate-fade-in-up transition-colors duration-300">
          <div class="sm:flex sm:items-start">
            <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 dark:bg-red-900/50 sm:mx-0 sm:h-10 sm:w-10">
              <svg class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            </div>
            <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
              <h3 class="text-lg leading-6 font-bold text-slate-900 dark:text-white" id="modal-title">
                Delete {{ userToDelete?.role }} Account
              </h3>
              <div class="mt-2 text-sm text-slate-500 dark:text-slate-400">
                Are you sure you want to permanently delete <b class="text-slate-700 dark:text-slate-300">{{ userToDelete?.full_name }}</b>? This action cannot be undone.
              </div>
            </div>
          </div>
          <div class="mt-6 flex justify-end space-x-3">
            <button @click="cancelDelete" type="button" class="px-4 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500 transition-colors">
              Cancel
            </button>
            <button @click="executeDelete" type="button" class="px-4 py-2 bg-red-600 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors">
              Delete Account
            </button>
          </div>
        </div>
      </div>
    </div>
    
    <AlertModal :show="alertState.show" :message="alertState.message" @close="alertState.show = false" />
  </div>
</template>
