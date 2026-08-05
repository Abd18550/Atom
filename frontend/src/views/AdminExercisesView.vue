<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import ConfirmModal from '../components/ConfirmModal.vue'
import AlertModal from '../components/AlertModal.vue'

const router = useRouter()
const exercises = ref([])
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
  type: 'program',
  right_solution: 'package main\n\nimport "fmt"\n\nfunc main() {\n\t// Write generic solution here\n}',
  test: 'package main\n\nimport "fmt"\n\nfunc main() {\n\t// Use sandbox/compare for assertions\n}'
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
  await fetchExercises()
})

const fetchExercises = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/exercises`)
    exercises.value = res.data
  } catch (err) {
    error.value = 'Failed to load exercises.'
  } finally {
    loading.value = false
  }
}

const openModal = async (exercise = null) => {
  if (exercise) {
    try {
      const res = await axios.get(`${API_BASE_URL}/api/exercises/${exercise.id}`)
      const fullEx = res.data
      editMode.value = true
      editId.value = fullEx.id
      form.value = {
        title: fullEx.title || '',
        description: fullEx.description || '',
        type: fullEx.type || 'program',
        right_solution: fullEx.right_solution || '',
        test: fullEx.test || ''
      }
    } catch (err) {
      alertState.value = { show: true, title: "Error", message: "Failed to fetch full exercise details." };
      return
    }
  } else {
    editMode.value = false
    editId.value = null
    form.value = {
      title: '',
      description: '',
      type: 'program',
      right_solution: 'package main\n\nimport "fmt"\n\nfunc main() {\n\t// Write correct solution here\n}',
      test: 'package main\n\nimport "fmt"\n\nfunc main() {\n\t// Test logic here\n}'
    }
  }
  submitError.value = ''
  submitSuccess.value = ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const addTestcase = () => {
  form.value.testcases.push('')
}

const removeTestcase = (index) => {
  if (form.value.testcases.length > 1) {
    form.value.testcases.splice(index, 1)
  }
}

const submitForm = async () => {
  submitError.value = ''
  submitSuccess.value = ''
  submitting.value = true
  
  if (!form.value.test.trim()) {
    submitError.value = "You must provide test code."
    submitting.value = false
    return
  }

  try {
    const payload = {
      title: form.value.title,
      description: form.value.description,
      type: form.value.type,
      right_solution: form.value.right_solution,
      test: form.value.test
    }

    if (editMode.value) {
      await axios.put(`${API_BASE_URL}/api/exercises/${editId.value}`, payload)
      submitSuccess.value = `Exercise updated successfully!`
    } else {
      await axios.post(`${API_BASE_URL}/api/exercises`, payload)
      submitSuccess.value = `Exercise created successfully!`
    }
    
    setTimeout(() => {
      closeModal()
      fetchExercises()
    }, 1500)
  } catch (err) {
    if (err.response && err.response.data && err.response.data.error) {
       submitError.value = err.response.data.error
    } else {
       submitError.value = 'Failed to save exercise.'
    }
  } finally {
    submitting.value = false
  }
}

const alertState = ref({ show: false, title: "Error", message: "" });
const confirmDelete = ref({ show: false, id: null, title: '' });

const deleteExercise = (id, title) => {
  confirmDelete.value = { show: true, id, title };
}

const executeDelete = async () => {
  const id = confirmDelete.value.id;
  confirmDelete.value.show = false;
  try {
    await axios.delete(`${API_BASE_URL}/api/exercises/${id}`)
    fetchExercises()
  } catch (err) {
    error.value = 'Failed to delete exercise.'
  }
}
</script>

<template>
  <div class="w-full max-w-6xl">
    <div class="mb-6 flex justify-between items-center transition-colors">
      <div>
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Manage Challenges</h2>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Create and monitor automated programming exercises</p>
      </div>
      <div>
        <button @click="openModal()" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-indigo-600 to-violet-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
          <svg class="-ml-1 mr-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          New Challenge
        </button>
      </div>
    </div>

    <!-- Exercises Table -->
    <div class="bg-white dark:bg-slate-800 shadow-md overflow-hidden sm:rounded-2xl border border-slate-200 dark:border-slate-700 transition-colors duration-300">
      <div v-if="loading" class="p-8 text-center text-slate-500 dark:text-slate-400">Loading exercises...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500 dark:text-red-400">{{ error }}</div>
      <table v-else class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
        <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700 transition-colors duration-300">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">ID</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Title</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Created At</th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-100 dark:divide-slate-700/50">
          <tr v-for="ex in exercises" :key="ex.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-slate-200">#{{ ex.id }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-indigo-600 dark:text-indigo-400">{{ ex.title }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-slate-400">
              {{ new Date(ex.created_at).toLocaleDateString() }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button @click="openModal(ex)" class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 mr-4 font-bold transition-colors">Edit</button>
              <button @click="deleteExercise(ex.id, ex.title)" class="text-red-600 dark:text-red-400 hover:text-red-900 dark:hover:text-red-300 font-bold transition-colors">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Exercise Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-900 bg-opacity-75 backdrop-blur-sm transition-opacity" @click="!submitting && closeModal()" aria-hidden="true"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl w-full border border-slate-100 dark:border-slate-700">
          <form @submit.prevent="submitForm" class="flex flex-col max-h-[90vh]">
            
            <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex justify-between items-center bg-slate-50 dark:bg-slate-900/50">
              <h3 class="text-xl font-bold text-slate-900 dark:text-white" id="modal-title">{{ editMode ? 'Edit Challenge' : 'Create New Challenge' }}</h3>
              <button type="button" @click="!submitting && closeModal()" class="text-slate-400 hover:text-slate-500 focus:outline-none">
                <span class="sr-only">Close</span>
                <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>

            <div class="px-6 py-6 overflow-y-auto flex-grow space-y-6">
              <div v-if="submitError" class="text-sm text-red-600 bg-red-50 dark:bg-red-900/30 dark:text-red-400 p-3 rounded-lg border border-red-200 dark:border-red-800">{{ submitError }}</div>
              <div v-if="submitSuccess" class="text-sm text-green-600 bg-green-50 dark:bg-green-900/30 dark:text-green-400 p-3 rounded-lg border border-green-200 dark:border-green-800">{{ submitSuccess }}</div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Left Stack: Title & Description -->
                <div class="space-y-6">
                  <div>
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Challenge Title</label>
                    <input v-model="form.title" type="text" required class="bg-white dark:bg-slate-900 dark:text-white block w-full border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2.5 px-3 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors">
                  </div>
                  <div>
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Instructions (Description)</label>
                    <textarea v-model="form.description" required rows="8" class="bg-white dark:bg-slate-900 dark:text-white block w-full border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2.5 px-3 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors font-mono text-sm leading-relaxed whitespace-pre-wrap"></textarea>
                  </div>
                </div>

                <!-- Right Stack: Right Solution Code -->
                <div class="h-full flex flex-col">
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Expected Solution (Go)</label>
                  <textarea v-model="form.right_solution" required spellcheck="false" class="flex-grow bg-slate-900 text-green-400 block w-full border border-slate-700 rounded-lg shadow-inner py-3 px-4 focus:outline-none focus:ring-2 focus:ring-green-500 sm:text-sm font-mono leading-relaxed whitespace-pre-wrap resize-none min-h-[300px]"></textarea>
                </div>
              </div>

              <!-- Test Code Section -->
              <div class="border-t border-slate-200 dark:border-slate-700 pt-6 space-y-4">
                <div class="flex items-center space-x-4 mb-4">
                  <label class="block text-lg font-bold text-slate-700 dark:text-slate-300">Challenge Type</label>
                  <select v-model="form.type" class="bg-white dark:bg-slate-900 dark:text-white border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2 px-3 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors">
                    <option value="program">Complete Program (os/exec via student/main.go)</option>
                    <option value="function">Function Only (import "sandbox/student")</option>
                  </select>
                </div>

                <div class="h-full flex flex-col">
                  <div class="flex justify-between items-center mb-1">
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300">Test Code / Custom Runner (Go)</label>
                    <span class="text-xs text-indigo-500 font-bold bg-indigo-50 px-2 py-1 rounded">Writes to test.go</span>
                  </div>
                  <textarea v-model="form.test" required spellcheck="false" class="flex-grow bg-slate-900 text-blue-400 block w-full border border-slate-700 rounded-lg shadow-inner py-3 px-4 focus:outline-none focus:ring-2 focus:ring-blue-500 sm:text-sm font-mono leading-relaxed whitespace-pre-wrap resize-none min-h-[300px]"></textarea>
                  <p class="mt-2 text-xs text-slate-500 dark:text-slate-400 italic">
                    <span v-if="form.type === 'program'">Use `os.Command` to run student binary and assert outputs using `Compare(expected, actual, caseName)`.</span>
                    <span v-if="form.type === 'function'">Import `"sandbox/student"`, call functions, and use `Compare(expected, actual, caseName)`.</span>
                    Remember to call `Success()` at the very end of main.
                  </p>
                </div>
              </div>

            </div>

            <div class="bg-gray-50 dark:bg-slate-900/80 px-6 py-4 border-t border-slate-200 dark:border-slate-700 flex justify-end space-x-3 transition-colors">
              <button type="button" @click="closeModal" :disabled="submitting" class="px-4 py-2 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 rounded-lg text-sm font-bold text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 focus:outline-none transition-colors disabled:opacity-50">
                Cancel
              </button>
              <button type="submit" :disabled="submitting" class="inline-flex items-center px-6 py-2 border border-transparent rounded-lg shadow-sm text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors disabled:opacity-50">
                <svg v-if="submitting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                {{ editMode ? 'Save Changes' : 'Publish Challenge' }}
              </button>
            </div>
            
          </form>
        </div>
      </div>
    </div>
    
    <!-- Custom Dialog Modals -->
    <AlertModal 
      :show="alertState.show" 
      :title="alertState.title" 
      :message="alertState.message" 
      @close="alertState.show = false" 
    />

    <ConfirmModal 
      :show="confirmDelete.show" 
      title="Delete Challenge" 
      :message="`Are you sure you want to completely delete &quot;${confirmDelete.title}&quot;? This will also remove all student submissions for it.`" 
      confirmText="Delete" 
      @confirm="executeDelete" 
      @cancel="confirmDelete.show = false" 
    />
  </div>
</template>
