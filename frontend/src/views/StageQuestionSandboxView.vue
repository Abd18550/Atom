<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'
import CodeEditor from '../components/CodeEditor.vue'

const route = useRoute()
const router = useRouter()

const exercise = ref(null)
const code = ref("")
const loading = ref(true)
const submitting = ref(false)
const submissionResult = ref(null)

const successModal = ref({
  show: false,
  title: '',
  message: '',
  encouragement: '',
  countdown: 5,
  nextQ: null,
  isLast: false
})
let pollingInterval = null
let countdownTimer = null

const encouragements = [
  "Great job! Perfect logic & syntax execution 🚀",
  "Awesome! Outstanding problem-solving skills 🌟",
  "Flawless code! Your Go mastery is advancing rapidly 👏",
  "Brilliant execution! Keep pushing forward 💻"
]

const goToNextQuestion = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  const target = successModal.value.nextTarget
  successModal.value.show = false
  if (target) {
    router.replace(target)
  } else {
    router.replace({ name: 'learning-path' })
  }
}

const fetchQuestion = async () => {
  loading.value = true
  submissionResult.value = null
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  successModal.value.show = false
  
  try {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_BASE_URL}/api/stage-questions/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    exercise.value = response.data
    // Suggest starting code depending on type
    if (exercise.value.type === 'function') {
      code.value = "package student\n\n// Write your function here\n"
    } else {
      code.value = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\t// Write your code here\n}\n"
    }
  } catch (err) {
    console.error("Failed to load stage question", err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchQuestion()
})

watch(() => route.params.id, (newId) => {
  if (newId && route.name === 'stage-question-sandbox') {
    fetchQuestion()
  }
})

const rateLimitCooldown = ref(0)
let cooldownTimer = null

const startRateLimitCooldown = (seconds = 5) => {
  rateLimitCooldown.value = seconds
  if (cooldownTimer) clearInterval(cooldownTimer)
  cooldownTimer = setInterval(() => {
    rateLimitCooldown.value--
    if (rateLimitCooldown.value <= 0) {
      clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
  if (countdownTimer) clearInterval(countdownTimer)
  if (cooldownTimer) clearInterval(cooldownTimer)
})

const submitCode = async () => {
  if (submitting.value || rateLimitCooldown.value > 0 || !code.value.trim()) return
  
  submitting.value = true
  startRateLimitCooldown(5)
  submissionResult.value = { status: 'Pending', message: 'Queued for grading...' }

  try {
    const token = localStorage.getItem('token')
    const response = await axios.post(`${API_BASE_URL}/api/submissions`, {
      stage_question_id: parseInt(route.params.id),
      code: code.value
    }, {
      headers: { Authorization: `Bearer ${token}` }
    })

    const submissionID = response.data.submission_id
    pollSubmissionStatus(submissionID)
  } catch (err) {
    console.error("Submission failed", err)
    let errMsg = err.response?.data?.error || 'Failed to submit code. Please try again.'
    if (err.response?.status === 429) {
      startRateLimitCooldown(5)
    }
    submissionResult.value = {
      status: 'Error',
      error_message: errMsg
    }
    submitting.value = false
  }
}

const handleSuccess = async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_BASE_URL}/api/learning-path`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    
    const stages = response.data.stages || response.data || []
    const currentQId = parseInt(route.params.id)
    
    let currentStageIndex = -1
    let currentQIndex = -1
    let currentStage = null
    
    for (let sIdx = 0; sIdx < stages.length; sIdx++) {
      const stage = stages[sIdx]
      const questions = stage.questions || stage.exercises || []
      const qIdx = questions.findIndex(q => q.id === currentQId)
      if (qIdx !== -1) {
        currentStage = stage
        currentStageIndex = sIdx
        currentQIndex = qIdx
        break
      }
    }

    if (currentStage && currentQIndex !== -1) {
      const stageQuestions = currentStage.questions || currentStage.exercises || []
      const nextQ = stageQuestions[currentQIndex + 1]
      const randomEncouragement = encouragements[Math.floor(Math.random() * encouragements.length)]

      if (countdownTimer) clearInterval(countdownTimer)

      if (nextQ) {
        // Next task in the same stage
        successModal.value = {
          show: true,
          badgeLabel: 'CHALLENGE COMPLETED',
          title: 'GOOD JOB!',
          encouragement: randomEncouragement,
          targetLabel: 'NEXT TASK',
          targetTitle: nextQ.title,
          countdown: 5,
          nextTarget: { name: 'stage-question-sandbox', params: { id: nextQ.id } },
          buttonText: 'Next Task',
          isLast: false
        }
      } else {
        // Last question in this stage -> Next Stage
        const nextStage = stages[currentStageIndex + 1]
        let nextTarget = { name: 'learning-path' }
        let targetLabel = 'STAGE COMPLETED'
        let targetTitle = `You've mastered "${currentStage.title}"!`

        if (nextStage) {
          const nextStageQuestions = nextStage.questions || nextStage.exercises || []
          if (nextStageQuestions.length > 0) {
            nextTarget = { name: 'stage-question-sandbox', params: { id: nextStageQuestions[0].id } }
            targetLabel = 'NEXT STAGE'
            targetTitle = nextStage.title
          } else {
            nextTarget = { name: 'stage-orbit', params: { id: nextStage.id } }
            targetLabel = 'NEXT STAGE'
            targetTitle = nextStage.title
          }
        }

        successModal.value = {
          show: true,
          badgeLabel: 'STAGE MASTERED 🏆',
          title: 'STAGE COMPLETED!',
          encouragement: 'Outstanding! You have mastered all challenges in this stage! 🌟',
          targetLabel: targetLabel,
          targetTitle: targetTitle,
          countdown: 5,
          nextTarget: nextTarget,
          buttonText: 'Next Stage',
          isLast: true
        }
      }

      countdownTimer = setInterval(() => {
        successModal.value.countdown--
        if (successModal.value.countdown <= 0) {
          goToNextQuestion()
        }
      }, 1000)
    } else {
      router.replace({ name: 'learning-path' })
    }
  } catch (e) {
    console.error("Failed to determine next navigation point", e)
    router.replace({ name: 'learning-path' })
  }
}

const pollSubmissionStatus = (id) => {
  if (pollingInterval) clearInterval(pollingInterval)
  
  pollingInterval = setInterval(async () => {
    try {
      const token = localStorage.getItem('token')
      const res = await axios.get(`${API_BASE_URL}/api/submissions/${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      
      const status = res.data.status
      if (status !== 'Pending' && status !== 'Running') {
        clearInterval(pollingInterval)
        submissionResult.value = res.data
        submitting.value = false
        
        if (status === 'Passed') {
          handleSuccess()
        }
      } else {
         submissionResult.value = { ...res.data, message: 'Running in Secure Sandbox...' }
      }
    } catch (err) {
      console.error("Polling error", err)
      clearInterval(pollingInterval)
      submitting.value = false
    }
  }, 1500)
}
</script>

<template>
  <div v-if="loading" class="text-white animate-pulse">Loading Question...</div>
  
  <div v-else-if="exercise" class="w-full max-w-7xl min-h-[calc(100vh-5rem)] lg:h-[calc(100vh-6rem)] flex flex-col lg:flex-row gap-6 mx-auto pb-12 lg:pb-0 overflow-y-auto lg:overflow-hidden px-2 sm:px-4">
    
    <!-- Left Panel: Description & Results -->
    <div class="w-full lg:w-1/3 flex flex-col gap-6 flex-shrink-0">
      
      <!-- Instructions Card -->
      <div class="bg-gray-800/80 backdrop-blur rounded-2xl p-4 sm:p-6 shadow-xl border border-gray-700/50 flex flex-col max-h-[380px] lg:max-h-none lg:flex-grow">
        <h1 class="text-xl sm:text-2xl font-bold text-white mb-2">{{ exercise.title }}</h1>
        <div class="prose prose-invert prose-indigo max-w-none text-gray-300 whitespace-pre-wrap flex-grow overflow-y-auto font-mono text-xs sm:text-sm leading-relaxed pr-1">
          {{ exercise.description }}
        </div>
      </div>

      <!-- Results Card (Shows when submitted) -->
      <div v-if="submissionResult" class="bg-gray-800/80 backdrop-blur rounded-2xl p-4 sm:p-6 shadow-xl border border-gray-700/50">
        <h2 class="text-lg font-bold text-white mb-4">Output Console</h2>
        
        <div class="space-y-4">
          <!-- Status Badge -->
          <div class="flex items-center space-x-2">
            <span class="font-semibold text-gray-300">Status:</span>
            <span :class="{
              'px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider': true,
              'bg-yellow-500/20 text-yellow-500': submissionResult.status === 'Pending' || submissionResult.status === 'Running',
              'bg-green-500/20 text-green-500': submissionResult.status === 'Passed',
              'bg-red-500/20 text-red-500': submissionResult.status === 'Failed' || submissionResult.status === 'SyntaxError' || submissionResult.status === 'Error'
            }">
              {{ submissionResult.status }}
            </span>
          </div>

          <!-- Pending Message -->
          <div v-if="submitting" class="text-indigo-400 font-mono text-sm animate-pulse">
            {{ submissionResult.message }}
          </div>

          <!-- Syntax / Runtime Error -->
          <div v-if="submissionResult.status === 'SyntaxError' || submissionResult.status === 'Error'" class="bg-red-900/30 border border-red-500/50 rounded-lg p-3 text-red-400 font-mono text-xs sm:text-sm whitespace-pre-wrap overflow-x-auto">
            {{ submissionResult.error_message || submissionResult.message }}
          </div>

          <!-- Logic Failure -->
          <div v-if="submissionResult.status === 'Failed'" class="space-y-3">
             <div class="text-xs sm:text-sm font-semibold text-red-400">Failed on test input: <span class="bg-gray-900 px-2 py-1 rounded">{{ submissionResult.failed_testcase }}</span></div>
             <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="text-xs text-gray-400 mb-1">Expected Output</div>
                  <pre class="bg-gray-900 text-green-400 p-2 rounded text-xs border border-gray-700 overflow-x-auto">{{ submissionResult.expected_output }}</pre>
                </div>
                <div>
                  <div class="text-xs text-gray-400 mb-1">Your Output</div>
                  <pre class="bg-gray-900 text-red-400 p-2 rounded text-xs border border-gray-700 overflow-x-auto">{{ submissionResult.actual_output }}</pre>
                </div>
             </div>
          </div>

          <!-- Success Message -->
          <div v-if="submissionResult.status === 'Passed'" class="text-green-400 font-mono text-sm">
            Challenge passed! You may return to the map.
          </div>
        </div>
      </div>
    </div>

    <!-- Right Panel: Code Editor -->
    <div class="w-full lg:w-2/3 min-h-[420px] sm:min-h-[500px] lg:min-h-0 bg-gray-900 rounded-2xl shadow-xl flex flex-col border border-gray-700/50 overflow-hidden flex-grow">
      <div class="bg-gray-800 px-4 py-3 flex justify-between items-center border-b border-gray-700/50">
         <div class="text-gray-300 font-semibold flex items-center gap-2">
            <svg class="h-5 w-5 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
            {{ exercise.type === 'function' ? 'solution.go' : 'main.go' }}
         </div>
         <button 
          @click="submitCode"
          :disabled="submitting || rateLimitCooldown > 0 || !code.trim()"
          class="flex items-center space-x-2 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-400 hover:to-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed text-white text-sm font-bold py-1.5 px-4 rounded-lg shadow-lg transition-all"
        >
          <svg v-if="submitting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <svg v-else class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ submitting ? 'Running...' : (rateLimitCooldown > 0 ? `Wait (${rateLimitCooldown}s)` : 'Run Code') }}</span>
        </button>
      </div>
      <div class="flex-grow relative overflow-hidden">
        <CodeEditor v-model="code" />
      </div>
    </div>

  </div>
  <div v-else class="text-red-400">Question not found.</div>

  <!-- Minimal Success Celebration Overlay (No Modal Box, Pure Translucent Overlay) -->
  <div v-if="successModal.show" class="fixed inset-0 z-[100] flex flex-col items-center justify-center p-6 bg-slate-950/75 backdrop-blur-lg text-center font-sans" dir="ltr">
    
    <!-- Animated Checkmark Icon -->
    <div class="mb-5 w-20 h-20 rounded-full bg-emerald-500/20 border border-emerald-400/50 flex items-center justify-center text-emerald-400 shadow-[0_0_60px_rgba(16,185,129,0.5)] animate-bounce">
      <svg class="h-10 w-10" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="3">
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path>
      </svg>
    </div>

    <!-- Minimal Headline -->
    <h2 class="text-5xl md:text-7xl font-black bg-gradient-to-r from-emerald-400 via-teal-300 to-cyan-400 bg-clip-text text-transparent tracking-tight mb-3 drop-shadow-xl">
      Great Job!
    </h2>

    <!-- Encouragement / Subtitle -->
    <p class="text-slate-200 text-lg md:text-xl font-medium max-w-lg mb-8 leading-relaxed drop-shadow">
      {{ successModal.encouragement || 'Challenge Passed Successfully!' }}
    </p>

    <!-- Floating Action Button -->
    <button 
      @click="goToNextQuestion" 
      class="bg-gradient-to-r from-emerald-500 via-teal-500 to-cyan-500 hover:from-emerald-400 hover:to-cyan-400 text-slate-950 font-black px-8 py-4 rounded-2xl shadow-[0_0_50px_rgba(16,185,129,0.45)] transition-all flex items-center gap-3 text-lg md:text-xl cursor-pointer transform hover:scale-105 mb-5"
    >
      <span>{{ successModal.buttonText || (successModal.isLast ? 'Next Stage' : 'Next Task') }}</span>
      <span class="bg-slate-950/30 text-slate-950 px-3 py-1 rounded-full text-sm font-extrabold font-mono">
        {{ successModal.countdown }}s
      </span>
      <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3"></path>
      </svg>
    </button>

    <!-- Secondary Link -->
    <button 
      @click="router.replace({ name: 'learning-path' })" 
      class="text-slate-400 hover:text-slate-200 font-bold text-sm tracking-wide transition-colors cursor-pointer"
    >
      Back to Learning Path
    </button>

  </div>
</template>
