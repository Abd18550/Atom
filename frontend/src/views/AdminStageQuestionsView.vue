<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import ConfirmModal from '../components/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()

const stageId = route.params.id
const stage = ref(null)
const questions = ref([])
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
  type: 'function',
  right_solution: 'package solution\n\nfunc MyFunction() {\n\t// Solution\n}',
  test: 'package main\n\nimport (\n\t"student"\n\t"solution"\n\t"testing"\n)\n\nfunc TestFunction(t *testing.T) {\n\t// Assertions\n}',
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
  await fetchData()
})

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE_URL}/api/stages`)
    const currentStage = res.data.find(s => s.id == stageId)
    if (!currentStage) {
      error.value = 'Stage not found.'
      return
    }
    stage.value = currentStage
    // Sort questions by order_index
    questions.value = (currentStage.questions || []).sort((a, b) => a.order_index - b.order_index)
  } catch (err) {
    error.value = 'Failed to load stage data.'
  } finally {
    loading.value = false
  }
}

const openModal = (q = null) => {
  if (q) {
    editMode.value = true
    editId.value = q.id
    form.value = {
      title: q.title || '',
      description: q.description || '',
      type: q.type || 'function',
      right_solution: q.right_solution || '',
      test: q.test || '',
      order_index: q.order_index || Math.max(1, questions.value.length)
    }
  } else {
    editMode.value = false
    editId.value = null
    form.value = {
      title: '',
      description: '',
      type: 'function',
      right_solution: 'package solution\n\nfunc MyFunction() {\n\t// Solution\n}',
      test: 'package main\n\nimport (\n\t"student"\n\t"solution"\n\t"testing"\n)\n\nfunc TestFunction(t *testing.T) {\n\t// Assertions\n}',
      order_index: questions.value.length + 1
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
      type: form.value.type,
      right_solution: form.value.right_solution,
      test: form.value.test,
      order_index: parseInt(form.value.order_index) || 1
    }

    if (editMode.value) {
      await axios.put(`${API_BASE_URL}/api/stages/${stageId}/questions/${editId.value}`, payload)
      submitSuccess.value = `Question updated successfully!`
    } else {
      await axios.post(`${API_BASE_URL}/api/stages/${stageId}/questions`, payload)
      submitSuccess.value = `Question created successfully!`
    }
    
    setTimeout(() => {
      closeModal()
      fetchData()
    }, 1500)
  } catch (err) {
    submitError.value = 'Failed to save question.'
  } finally {
    submitting.value = false
  }
}

const confirmDelete = ref({ show: false, id: null, title: '' });

const deleteQuestion = (id, title) => {
  confirmDelete.value = { show: true, id, title };
}

const executeDelete = async () => {
  const id = confirmDelete.value.id;
  confirmDelete.value.show = false;
  try {
    await axios.delete(`${API_BASE_URL}/api/stages/${stageId}/questions/${id}`)
    fetchData()
  } catch (err) {
    error.value = 'Failed to delete question.'
  }
}
</script>

<template>
  <div class="w-full max-w-6xl">
    <div class="mb-6 flex justify-between items-center">
      <div v-if="stage">
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Stage {{ stage.order_index }}: {{ stage.title }}</h2>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage the coding challenges for this stage</p>
      </div>
      <div v-else>
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Loading...</h2>
      </div>
      <div>
        <button v-if="stage" @click="openModal()" class="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm font-bold text-white bg-gradient-to-r from-sky-500 to-indigo-600 focus:outline-none">
          + New Question
        </button>
      </div>
    </div>

    <!-- Questions Table -->
    <div class="bg-white dark:bg-slate-800 shadow-md overflow-hidden sm:rounded-2xl border border-slate-200 dark:border-slate-700 transition-colors duration-300">
      <div v-if="loading" class="p-8 text-center text-slate-500 dark:text-slate-400">Loading details...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500 dark:text-red-400">{{ error }}</div>
      <div v-else-if="questions.length === 0" class="p-8 text-center text-slate-500 dark:text-slate-400">No questions found for this stage yet. Create one!</div>
      <table v-else class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
        <thead class="bg-slate-50 dark:bg-slate-900/50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Order</th>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Title</th>
            <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Type</th>
            <th class="px-6 py-3 text-right text-xs font-bold text-slate-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-100 dark:divide-slate-700/50">
          <tr v-for="q in questions" :key="q.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-sky-500"># {{ q.order_index }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-slate-900 dark:text-white">{{ q.title }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500 dark:text-slate-400 uppercase font-mono">{{ q.type }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button @click="openModal(q)" class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 mr-4 font-bold">Edit</button>
              <button @click="deleteQuestion(q.id, q.title)" class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 font-bold">Delete</button>
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

        <div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle w-full max-w-4xl border border-slate-100 dark:border-slate-700">
          <form @submit.prevent="submitForm">
            <div class="px-6 pt-6 pb-6 h-[80vh] overflow-y-auto">
              <h3 class="text-2xl font-bold text-slate-900 dark:text-white mb-6 border-b border-slate-200 dark:border-slate-700 pb-3">{{ editMode ? 'Edit Question' : 'Create New Question' }}</h3>
              
              <div v-if="submitError" class="text-sm text-red-600 mb-4 p-3 bg-red-50 dark:bg-red-900/30 rounded-lg">{{ submitError }}</div>
              <div v-if="submitSuccess" class="text-sm text-green-600 mb-4 p-3 bg-green-50 dark:bg-green-900/30 rounded-lg">{{ submitSuccess }}</div>

              <div class="grid grid-cols-1 gap-6">
                <div class="grid grid-cols-3 gap-6">
                  <div class="col-span-2">
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Question Title</label>
                    <input v-model="form.title" type="text" required class="w-full px-4 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white focus:ring-2 focus:ring-indigo-500 outline-none">
                  </div>
                  <div>
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Order Index</label>
                    <input v-model="form.order_index" type="number" min="1" required class="w-full px-4 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white focus:ring-2 focus:ring-indigo-500 outline-none">
                  </div>
                </div>
                
                <div>
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Instructions (with Examples)</label>
                  <textarea v-model="form.description" required rows="6" class="w-full px-4 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white font-mono text-sm leading-relaxed focus:ring-2 focus:ring-indigo-500 outline-none"></textarea>
                  <p class="mt-1 text-xs text-slate-500">Provide clear instructions and include an 'Examples:' section.</p>
                </div>

                <div>
                  <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Type</label>
                  <select v-model="form.type" class="w-full px-4 py-2 border rounded-lg dark:bg-slate-900 dark:border-slate-600 dark:text-white focus:ring-2 focus:ring-indigo-500 outline-none">
                    <option value="program">Program (Standard I/O)</option>
                    <option value="function">Function (Unit Test Binding)</option>
                  </select>
                </div>

                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Right Solution Code</label>
                    <textarea v-model="form.right_solution" required rows="14" class="w-full px-4 py-2 border rounded-lg bg-slate-50 dark:bg-slate-900 dark:border-slate-600 dark:text-sky-300 text-sky-700 font-mono text-sm leading-relaxed focus:ring-2 focus:ring-indigo-500 outline-none whitespace-pre"></textarea>
                    <p class="mt-1 text-xs text-slate-500">Use `package solution` for function types.</p>
                  </div>
                  <div>
                    <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1">Test Code</label>
                    <textarea v-model="form.test" required rows="14" class="w-full px-4 py-2 border rounded-lg bg-slate-50 dark:bg-slate-900 dark:border-slate-600 dark:text-emerald-300 text-emerald-700 font-mono text-sm leading-relaxed focus:ring-2 focus:ring-indigo-500 outline-none whitespace-pre"></textarea>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="bg-slate-50 dark:bg-slate-900/80 px-6 py-4 flex justify-end space-x-3 border-t border-slate-200 dark:border-slate-700">
              <button type="button" @click="closeModal" class="px-5 py-2.5 bg-white border border-slate-300 rounded-lg text-slate-700 dark:bg-slate-800 dark:text-white dark:border-slate-600 font-bold hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
              <button type="submit" :disabled="submitting" class="px-5 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-bold disabled:opacity-50 transition-colors">
                {{ editMode ? 'Save Changes' : 'Create Question' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
    
    <!-- Custom Confirm Modal -->
    <ConfirmModal 
      :show="confirmDelete.show" 
      title="Delete Question" 
      :message="`Are you sure you want to completely delete &quot;${confirmDelete.title}&quot;?`" 
      confirmText="Delete" 
      @confirm="executeDelete" 
      @cancel="confirmDelete.show = false" 
    />
  </div>
</template>
