<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import ConfirmModal from '../components/ConfirmModal.vue'

const router = useRouter()
const stages = ref([])
const loading = ref(true)
const error = ref('')

const showModal = ref(false)
const submitError = ref('')
const submitSuccess = ref('')
const submitting = ref(false)
const editMode = ref(false)
const editId = ref(null)

const form = ref({
  title: '',
  description: '',
  order_index: 0
})

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
  await fetchStages()
})

const fetchStages = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/stages`)
    stages.value = res.data
  } catch (err) {
    error.value = 'Failed to load stages.'
  } finally {
    loading.value = false
  }
}

const openModal = (stage = null) => {
  if (stage) {
    editMode.value = true
    editId.value = stage.id
    form.value = {
      title: stage.title || '',
      description: stage.description || '',
      order_index: stage.order_index || 0
    }
  } else {
    editMode.value = false
    editId.value = null
    form.value = {
      title: '',
      description: '',
      order_index: stages.value.length + 1
    }
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
  submitting.value = true

  try {
    const payload = {
      title: form.value.title,
      description: form.value.description,
      order_index: parseInt(form.value.order_index)
    }

    if (editMode.value) {
      await axios.put(`${API_BASE_URL}/api/stages/${editId.value}`, payload)
      submitSuccess.value = `Stage updated successfully!`
    } else {
      await axios.post(`${API_BASE_URL}/api/stages`, payload)
      submitSuccess.value = `Stage created successfully!`
    }
    
    setTimeout(() => {
      closeModal()
      fetchStages()
    }, 1500)
  } catch (err) {
    submitError.value = 'Failed to save stage.'
  } finally {
    submitting.value = false
  }
}

const confirmDelete = ref({ show: false, id: null, title: '' });

const deleteStage = (id, title) => {
  confirmDelete.value = { show: true, id, title };
}

const executeDelete = async () => {
  const id = confirmDelete.value.id;
  confirmDelete.value.show = false;
  try {
    await axios.delete(`${API_BASE_URL}/api/stages/${id}`)
    fetchStages()
  } catch (err) {
    error.value = 'Failed to delete stage.'
  }
}
</script>

<template>
  <div class="w-full max-w-6xl">
    <div class="mb-6 flex justify-between items-center">
      <div>
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Manage Learning Stages</h2>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Create and monitor the Go learning journey nodes</p>
      </div>
      <div>
        <button @click="openModal()" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-sky-500 to-indigo-600 focus:outline-none">
          <svg class="-ml-1 mr-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          New Stage
        </button>
      </div>
    </div>

    <!-- Stages Table -->
    <div class="bg-white dark:bg-slate-800 shadow-md overflow-hidden sm:rounded-2xl border border-slate-200 dark:border-slate-700 transition-colors duration-300">
      <div v-if="loading" class="p-8 text-center text-slate-500 dark:text-slate-400">Loading stages...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500 dark:text-red-400">{{ error }}</div>
      <table v-else class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
        <thead class="bg-slate-50 dark:bg-slate-900/50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Order</th>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Title</th>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Description</th>
            <th class="px-6 py-3 text-right text-xs font-bold text-slate-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-100 dark:divide-slate-700/50">
          <tr v-for="stage in stages" :key="stage.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-sky-500">Node {{ stage.order_index }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-slate-900 dark:text-white">{{ stage.title }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500 dark:text-slate-400">{{ stage.description }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <router-link :to="'/admin/stages/' + stage.id + '/questions'" class="text-sky-600 hover:text-sky-900 dark:text-sky-400 dark:hover:text-sky-300 mr-4 font-bold">Manage Questions</router-link>
              <button @click="openModal(stage)" class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 mr-4 font-bold">Edit</button>
              <button @click="deleteStage(stage.id, stage.title)" class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 font-bold">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-900 bg-opacity-75 backdrop-blur-sm transition-opacity" @click="!submitting && closeModal()" aria-hidden="true"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg w-full border border-slate-100 dark:border-slate-700">
          <form @submit.prevent="submitForm">
            <div class="px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
              <h3 class="text-xl font-bold text-slate-900 dark:text-white mb-4 border-b border-slate-200 dark:border-slate-700 pb-3">{{ editMode ? 'Edit Stage' : 'Create New Stage' }}</h3>
              
              <div v-if="submitError" class="text-sm text-red-600 mb-4 p-2 bg-red-50 dark:bg-red-900/30 rounded">{{ submitError }}</div>
              <div v-if="submitSuccess" class="text-sm text-green-600 mb-4 p-2 bg-green-50 dark:bg-green-900/30 rounded">{{ submitSuccess }}</div>

              <div class="space-y-4">
                <div>
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Title</label>
                  <input v-model="form.title" type="text" required class="w-full px-3 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white">
                </div>
                <div>
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Description</label>
                  <textarea v-model="form.description" required rows="3" class="w-full px-3 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white"></textarea>
                </div>
                <div>
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Order Index (1 = First)</label>
                  <input v-model="form.order_index" type="number" min="1" required class="w-full px-3 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white">
                </div>
              </div>
            </div>
            
            <div class="bg-gray-50 dark:bg-slate-900/80 px-4 py-3 flex justify-end space-x-3">
              <button type="button" @click="closeModal" class="px-4 py-2 bg-white border border-gray-300 rounded-md text-gray-700 dark:bg-slate-700 dark:text-white dark:border-slate-500 font-medium">Cancel</button>
              <button type="submit" :disabled="submitting" class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 font-medium disabled:opacity-50 flex items-center">
                {{ editMode ? 'Save Changes' : 'Create Stage' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
    
    <!-- Custom Confirm Modal -->
    <ConfirmModal 
      :show="confirmDelete.show" 
      title="Delete Stage" 
      :message="`Are you sure you want to completely delete &quot;${confirmDelete.title}&quot;?`" 
      confirmText="Delete" 
      @confirm="executeDelete" 
      @cancel="confirmDelete.show = false" 
    />
  </div>
</template>
