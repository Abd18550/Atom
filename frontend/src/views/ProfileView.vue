<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const user = ref({})
const isEditing = ref(false)
const form = ref({
  email: '',
  current_password: '',
  new_password: '',
  confirm_password: '',
  avatar: ''
})
const availableRobots = ['Atom1', 'Atom2', 'Atom3', 'Atom4', 'Atom5', 'Atom6', 'Atom7', 'Atom8']
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    form.value.email = user.value.email
    form.value.avatar = user.value.avatar || ''
  }
})

const toggleEdit = () => {
  isEditing.value = !isEditing.value
  successMessage.value = ''
  errorMessage.value = ''
  if (!isEditing.value) {
    form.value.current_password = ''
    form.value.new_password = ''
    form.value.confirm_password = ''
    form.value.email = user.value.email
    form.value.avatar = user.value.avatar || ''
  } else {
    form.value.avatar = user.value.avatar || ''
  }
}

const updateProfile = async () => {
  successMessage.value = ''
  errorMessage.value = ''

  if (form.value.new_password && form.value.new_password !== form.value.confirm_password) {
    errorMessage.value = 'New passwords do not match.'
    return
  }

  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const res = await axios.put(`${API_BASE_URL}/api/profile`, {
      email: form.value.email,
      current_password: form.value.current_password,
      new_password: form.value.new_password,
      avatar: form.value.avatar
    }, {
      headers: { Authorization: `Bearer ${token}` }
    })

    successMessage.value = 'Profile updated successfully!'
    
    // Update local user state
    user.value = res.data.user
    localStorage.setItem('user', JSON.stringify(res.data.user))

    // Clear password fields
    form.value.current_password = ''
    form.value.new_password = ''
    form.value.confirm_password = ''
    
    setTimeout(() => {
        isEditing.value = false
        successMessage.value = ''
    }, 2000)

  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Failed to update profile'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-3xl animate-fade-in-up">
    <div class="bg-white dark:bg-slate-800 shadow-xl rounded-2xl overflow-hidden transition-colors duration-300">
      <!-- Header Banner (Cover Photo Style) -->
      <div class="h-40 bg-gradient-to-tr from-indigo-600 via-purple-600 to-violet-700 relative overflow-hidden">
        <div class="absolute inset-0 bg-white/10 blur-2xl"></div>
      </div>
      
      <!-- Content Area -->
      <div class="px-8 pb-10 sm:px-10">
        <!-- Overlapping Avatar -->
        <div class="flex justify-center -mt-16 mb-6">
            <div class="h-32 w-32 rounded-full bg-slate-50 dark:bg-slate-900 flex items-center justify-center text-indigo-700 dark:text-indigo-400 font-extrabold text-5xl uppercase border-8 border-white dark:border-slate-800 shadow-xl relative z-10 hover:scale-105 transition-transform duration-300 overflow-hidden">
                <img v-if="user.avatar" :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + user.avatar" class="w-full h-full object-cover bg-indigo-50 dark:bg-indigo-900" />
                <span v-else>{{ user.full_name ? user.full_name.charAt(0) : 'U' }}</span>
            </div>
        </div>

        <div class="text-center mb-8">
            <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white transition-colors">{{ user.full_name }}</h2>
            <p class="text-slate-500 dark:text-slate-400 font-medium mt-1 transition-colors">@{{ user.username }}</p>
        </div>

        <ul class="space-y-6 divide-y divide-slate-100 dark:divide-slate-700/50 max-w-xl mx-auto bg-slate-50 dark:bg-slate-900/50 rounded-2xl border border-slate-100 dark:border-slate-700/50 p-6 shadow-sm transition-colors duration-300">
            <li class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between first:pt-0">
                <span class="text-sm font-semibold text-slate-500 dark:text-slate-400 w-1/3">Email Address</span> 
                <span class="text-base font-medium text-slate-900 dark:text-slate-200 w-2/3">{{ user.email }}</span>
            </li>
            <li class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between">
                <span class="text-sm font-semibold text-slate-500 w-1/3">Role</span> 
                <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider w-fit shadow-sm"
                        :class="{
                        'bg-red-100 text-red-800': user.role === 'Admin',
                        'bg-purple-100 text-purple-800': user.role === 'Supervisor',
                        'bg-blue-100 text-blue-800': user.role === 'Mentor',
                        'bg-emerald-100 text-emerald-800': user.role === 'Student'
                        }">
                    {{ user.role }}
                </span>
            </li>
            <li v-if="user.school_name" class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between">
                <span class="text-sm font-semibold text-slate-500 dark:text-slate-400 w-1/3">School Name</span> 
                <span class="text-base font-medium text-slate-900 dark:text-slate-200 w-2/3">{{ user.school_name }}</span>
            </li>
            <li v-if="user.class" class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between">
                <span class="text-sm font-semibold text-slate-500 dark:text-slate-400 w-1/3">Class</span> 
                <span class="text-base font-medium text-slate-900 dark:text-slate-200 w-2/3">{{ user.class }}</span>
            </li>
            <li v-if="user.date_of_birth" class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between">
                <span class="text-sm font-semibold text-slate-500 dark:text-slate-400 w-1/3">Date of Birth</span> 
                <span class="text-base font-medium text-slate-900 dark:text-slate-200 w-2/3">{{ user.date_of_birth }}</span>
            </li>
        </ul>

        <div class="mt-10 max-w-xl mx-auto">
            <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-bold text-slate-900 dark:text-white transition-colors">Security & Settings</h3>
                <button @click="toggleEdit" class="text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 transition-colors">
                    {{ isEditing ? 'Cancel Edit' : 'Edit Profile' }}
                </button>
            </div>

            <div v-if="isEditing" class="bg-white dark:bg-slate-800/80 p-6 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-lg animate-fade-in-up transition-colors duration-300">
                <form @submit.prevent="updateProfile" class="space-y-5">
                    
                    <div v-if="successMessage" class="p-3 bg-emerald-50 text-emerald-700 text-sm rounded-lg border border-emerald-100">
                        {{ successMessage }}
                    </div>
                    <div v-if="errorMessage" class="p-3 bg-red-50 text-red-700 text-sm rounded-lg border border-red-100">
                        {{ errorMessage }}
                    </div>

                    <div>
                        <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-1">Email Address</label>
                        <input v-model="form.email" type="email" required class="w-full bg-transparent border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2 px-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 dark:text-white transition-all outline-none">
                    </div>

                    <div class="pt-4 border-t border-slate-100 dark:border-slate-700">
                        <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">Choose Your Avatar</label>
                        <div class="grid grid-cols-4 sm:grid-cols-4 gap-4 justify-items-center">
                            <div v-for="robot in availableRobots" :key="robot"
                                 @click="form.avatar = robot"
                                 :class="{'ring-4 ring-indigo-500 scale-110 shadow-lg': form.avatar === robot, 'hover:scale-105 hover:shadow-md cursor-pointer': true}"
                                 class="w-16 h-16 sm:w-20 sm:h-20 rounded-full bg-indigo-50 dark:bg-indigo-900/50 p-1 transition-all duration-200">
                                <img :src="'https://api.dicebear.com/7.x/bottts/svg?seed=' + robot" class="w-full h-full rounded-full" />
                            </div>
                        </div>
                    </div>

                    <div class="pt-2 border-t border-slate-100 dark:border-slate-700">
                        <p class="text-xs text-slate-500 dark:text-slate-400 mb-3 font-medium uppercase tracking-wider">Change Password</p>
                        
                        <div class="space-y-4">
                            <div>
                                <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-1">Current Password <span class="text-slate-400 dark:text-slate-500 font-normal">(Required to change password)</span></label>
                                <input v-model="form.current_password" type="password" class="w-full bg-transparent border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2 px-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 dark:text-white transition-all outline-none">
                            </div>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-1">New Password</label>
                                    <input v-model="form.new_password" type="password" class="w-full bg-transparent border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2 px-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 dark:text-white transition-all outline-none">
                                </div>
                                <div>
                                    <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-1">Confirm New Password</label>
                                    <input v-model="form.confirm_password" type="password" class="w-full bg-transparent border border-slate-300 dark:border-slate-600 rounded-lg shadow-sm py-2 px-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 dark:text-white transition-all outline-none">
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pt-4 flex justify-end">
                        <button type="submit" :disabled="loading" class="px-6 py-2.5 bg-gradient-to-r from-indigo-600 to-violet-600 text-white font-bold rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all text-sm disabled:opacity-70 disabled:cursor-not-allowed">
                            {{ loading ? 'Saving...' : 'Save Changes' }}
                        </button>
                    </div>
                </form>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up { animation: fadeInUp 0.5s ease-out forwards; }
</style>
