<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const router = useRouter()
const route = useRoute()
const stage = ref(null)
const loading = ref(true)
const hoveredNode = ref(null)

onMounted(async () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')

  if (!token || !userStr) {
    router.push('/login')
    return
  }

  try {
    const response = await axios.get(`${API_BASE_URL}/api/learning-path`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    
    // Find the specific stage
    const pathFlow = response.data || []
    stage.value = pathFlow.find(s => s.id == route.params.id)
    
    if (!stage.value) {
      router.push('/learning-path')
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
})

const isQuestionUnlocked = (idx) => {
  if (!stage.value) return false;
  if (!stage.value.unlocked) return false;
  if (idx === 0) return true;
  return stage.value.exercises[idx - 1].passed;
}

const getOverallStageStatus = () => {
  if (!stage.value) return 'locked';
  if (!stage.value.unlocked) return 'locked';
  if (stage.value.exercises && stage.value.exercises.length > 0 && stage.value.exercises.every(ex => ex.passed)) {
    return 'completed';
  }
  return 'active';
}

const getQuestionStatus = (ex, idx) => {
  if (ex.passed) return 'completed';
  if (isQuestionUnlocked(idx)) return 'active';
  return 'locked';
}

const startQuestion = (ex, idx) => {
  if (isQuestionUnlocked(idx)) {
    router.push('/stage-questions/' + ex.id)
  }
}

const layoutNodes = computed(() => {
  if (!stage.value || !stage.value.exercises) return [];
  const total = stage.value.exercises.length;
  if (total === 0) return [];
  if (total === 1) return [{ ...stage.value.exercises[0], originalIndex: 0, topPercent: '50%', leftPercent: '50%', tooltipPos: 'top-full mt-3 left-1/2 -translate-x-1/2' }];

  const startAngle = Math.PI / 2;
  const spread = (Math.PI * 2) / total;

  return stage.value.exercises.map((ex, i) => {
    const angle = startAngle + i * spread;
    const cx = 50, cy = 50, r = 40; 
    
    const dx = Math.cos(angle);
    const dy = Math.sin(angle);
    
    const left = cx + r * dx;
    const top = cy + r * dy;
    
    let tooltipPos = '';
    if (Math.abs(dx) >= 0.35) {
      if (dx > 0) {
        tooltipPos = 'left-full ml-3 top-1/2 -translate-y-1/2 origin-left';
      } else {
        tooltipPos = 'right-full mr-3 top-1/2 -translate-y-1/2 origin-right';
      }
    } else {
      if (dy > 0) {
        tooltipPos = 'top-full mt-3 left-1/2 -translate-x-1/2 origin-top';
      } else {
        tooltipPos = 'bottom-full mb-3 left-1/2 -translate-x-1/2 origin-bottom';
      }
    }
    
    return { 
      ...ex, 
      originalIndex: i,
      topPercent: `${top}%`,
      leftPercent: `${left}%`, 
      tooltipPos
    };
  })
})
</script>

<template>
  <div class="relative w-full h-[calc(100vh-80px)] bg-slate-50 dark:bg-slate-900 transition-colors duration-300 font-sans overflow-hidden flex flex-col items-center">
    
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
       <div class="absolute top-[20%] left-[20%] w-[20rem] h-[20rem] bg-indigo-500/10 dark:bg-indigo-600/10 rounded-full blur-[80px]"></div>
       <div class="absolute bottom-[10%] right-[20%] w-[25rem] h-[25rem] bg-emerald-500/10 dark:bg-emerald-600/10 rounded-full blur-[80px]"></div>
    </div>

    <!-- Header -->
    <div class="text-center pt-8 z-10 relative flex-shrink-0 w-full mb-2">
      <h1 class="text-3xl md:text-5xl font-black text-slate-900 dark:text-white tracking-tight drop-shadow-sm px-16">
        {{ stage ? stage.title : 'Loading...' }}
      </h1>
      <p class="text-slate-600 dark:text-slate-400 text-xs md:text-sm font-medium mt-2 px-16">
         {{ stage ? stage.description : '...' }}
      </p>
    </div>

    <div v-if="loading" class="flex-grow flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold animate-pulse text-xl z-10">
      Locating Questions...
    </div>

    <!-- Geometric Container -->
    <div v-else class="flex-grow w-full flex items-center justify-center z-10 my-auto p-4">
      <div class="relative w-full" style="max-width: min(100%, 70vh, 800px); aspect-ratio: 1/1;">
        
        <svg viewBox="0 0 100 100" class="absolute inset-0 w-full h-full pointer-events-none drop-shadow-sm overflow-visible">
          <!-- Orbit -->
          <circle cx="50" cy="50" r="40" class="stroke-indigo-400/30 dark:stroke-indigo-400/30" stroke-width="0.2" stroke-dasharray="1 2" fill="none" />
          
          <!-- Nucleus -->
          <circle cx="50" cy="50" r="2.5" class="animate-pulse" 
            :class="getOverallStageStatus() === 'completed' ? 'fill-emerald-500/40 dark:fill-emerald-400/40' : 'fill-indigo-500/40 dark:fill-indigo-400/40'" />
          <circle cx="50" cy="50" r="6" class="stroke-indigo-400/20 dark:stroke-indigo-300/20" stroke-width="0.1" fill="none" />

          <!-- Forward Lines -->
          <line 
            v-for="(node, i) in layoutNodes.slice(0, -1)" 
            :key="'line-'+i"
            :x1="parseFloat(node.leftPercent)" 
            :y1="parseFloat(node.topPercent)" 
            :x2="parseFloat(layoutNodes[i+1].leftPercent)" 
            :y2="parseFloat(layoutNodes[i+1].topPercent)" 
            class="stroke-slate-300/80 dark:stroke-slate-600 transition-colors duration-300"
            stroke-width="0.4"
          />
          
          <!-- Loop Line -->
          <line 
            v-if="layoutNodes.length > 2"
            :x1="parseFloat(layoutNodes[layoutNodes.length - 1].leftPercent)" 
            :y1="parseFloat(layoutNodes[layoutNodes.length - 1].topPercent)" 
            :x2="parseFloat(layoutNodes[0].leftPercent)" 
            :y2="parseFloat(layoutNodes[0].topPercent)" 
            class="stroke-slate-300/80 dark:stroke-slate-600 transition-colors duration-300"
            stroke-width="0.4"
          />
        </svg>

        <!-- Nodes -->
        <div 
          v-for="node in layoutNodes" 
          :key="node.id"
          class="absolute transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer"
          :style="{ top: node.topPercent, left: node.leftPercent }"
          @click="startQuestion(node, node.originalIndex)"
        >
          <div class="relative flex flex-col items-center justify-center transition-transform duration-300 group-hover:scale-110">
             
             <div class="relative flex items-center justify-center">
               <div v-if="getQuestionStatus(node, node.originalIndex) !== 'locked'" 
                    class="absolute inset-0 rounded-full animate-ping opacity-20"
                    :class="getQuestionStatus(node, node.originalIndex) === 'completed' ? 'bg-emerald-400' : 'bg-indigo-400'">
               </div>
               
               <div 
                 class="relative z-10 w-10 h-10 md:w-12 md:h-12 flex items-center justify-center shadow-lg border-2 transition-all duration-300 backdrop-blur-md rounded-full"
                 :class="[
                    getQuestionStatus(node, node.originalIndex) === 'locked' ? 'bg-slate-200/80 border-slate-300 dark:bg-slate-800/80 dark:border-slate-700 text-slate-400 dark:text-slate-600' :
                    getQuestionStatus(node, node.originalIndex) === 'completed' ? 'bg-emerald-100 border-emerald-400 text-emerald-500 dark:bg-emerald-900/60 dark:border-emerald-500 shadow-[0_0_15px_rgba(16,185,129,0.4)]' :
                    'bg-indigo-100 border-indigo-500 text-indigo-600 dark:bg-indigo-900/60 dark:text-indigo-400 dark:border-indigo-400 shadow-[0_0_15px_rgba(99,102,241,0.5)]'
                 ]"
               >
                  <svg v-if="getQuestionStatus(node, node.originalIndex) === 'locked'" class="w-4 h-4 md:w-5 md:h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" fill-rule="evenodd"></path></svg>
                  <svg v-else-if="getQuestionStatus(node, node.originalIndex) === 'completed'" class="w-5 h-5 md:w-6 md:h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
                  <span v-else class="font-bold text-lg md:text-xl">{{ node.originalIndex + 1 }}</span>
               </div>
             </div>

             <!-- Task Title Pop-up on the RADIAL OUTER side of Node on Hover (Non-overlapping, Compact & Truncated) -->
             <div 
               class="absolute w-max max-w-[130px] sm:max-w-[160px] pointer-events-none opacity-0 group-hover:opacity-100 scale-90 group-hover:scale-100 transition-all duration-200 ease-out z-50"
               :class="node.tooltipPos"
             >
               <div 
                 class="px-2.5 py-1 text-[11px] font-extrabold tracking-tight rounded-xl bg-slate-950/95 text-slate-100 border border-slate-700/80 shadow-2xl backdrop-blur-md text-center truncate"
                 :class="getQuestionStatus(node, node.originalIndex) === 'locked' ? 'text-slate-400 border-slate-800' : 'text-indigo-200 border-indigo-500/50 shadow-[0_0_15px_rgba(99,102,241,0.25)]'"
               >
                 {{ node.title }}
               </div>
             </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>
