<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL } from '../config.js'

const router = useRouter()
const pathFlow = ref([])
const loading = ref(true)
const hoveredStage = ref(null)

// Modal state
const showModal = ref(false)
const selectedStage = ref(null)

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
    return
  }
  try {
    const response = await axios.get(`${API_BASE_URL}/api/learning-path`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    pathFlow.value = response.data
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
})

const isAllExercisesPassed = (stage) => {
  if (!stage.exercises || stage.exercises.length === 0) return false;
  return stage.exercises.every(ex => ex.passed);
}

const getStageStatus = (stage) => {
  if (!stage.unlocked) return 'locked';
  if (isAllExercisesPassed(stage)) return 'completed';
  return 'active';
}

const startStage = (stage) => {
  if (getStageStatus(stage) !== 'locked') {
    router.push('/stage/' + stage.id)
  }
}

// Generates a mathematically precise orbital layout
const layoutNodes = computed(() => {
  const total = pathFlow.value.length;
  if (total === 0) return [];
  if (total === 1) return [{ ...pathFlow.value[0], topPercent: '50%', leftPercent: '50%', tooltipPos: 'top-full mt-3 left-1/2 -translate-x-1/2' }];

  // Mathematically distribute along a full circle evenly (Clock System)
  const startAngle = Math.PI / 2;
  const spread = (Math.PI * 2) / total;

  return pathFlow.value.map((stage, i) => {
    const angle = startAngle + i * spread;
    
    const cx = 50;
    const cy = 50;
    const r = 40; 
    
    const dx = Math.cos(angle);
    const dy = Math.sin(angle);
    
    const left = cx + r * dx;
    const top = cy + r * dy;
    
    // Determine outer direction relative to orbit center (50, 50)
    let tooltipPos = '';
    if (Math.abs(dx) >= 0.35) {
      if (dx > 0) {
        // Outer Right side of node
        tooltipPos = 'left-full ml-3 top-1/2 -translate-y-1/2 origin-left';
      } else {
        // Outer Left side of node
        tooltipPos = 'right-full mr-3 top-1/2 -translate-y-1/2 origin-right';
      }
    } else {
      if (dy > 0) {
        // Outer Bottom side of node
        tooltipPos = 'top-full mt-3 left-1/2 -translate-x-1/2 origin-top';
      } else {
        // Outer Top side of node
        tooltipPos = 'bottom-full mb-3 left-1/2 -translate-x-1/2 origin-bottom';
      }
    }
    
    return { 
      ...stage, 
      topPercent: `${top}%`,
      leftPercent: `${left}%`,
      tooltipPos
    };
  })
})
</script>

<template>
  <div class="relative w-full h-[calc(100vh-80px)] bg-slate-50 dark:bg-slate-900 transition-colors duration-300 font-sans overflow-hidden flex flex-col items-center">
    
    <!-- Background Decorators / Ambient Glare -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
       <div class="absolute top-[20%] left-[20%] w-[20rem] h-[20rem] bg-indigo-500/10 dark:bg-indigo-600/10 rounded-full blur-[80px]"></div>
       <div class="absolute bottom-[10%] right-[20%] w-[25rem] h-[25rem] bg-pink-500/10 dark:bg-pink-600/10 rounded-full blur-[80px]"></div>
    </div>

    <!-- Header -->
    <div class="text-center pt-8 z-10 relative flex-shrink-0">
      <h1 class="text-3xl md:text-5xl font-black text-slate-900 dark:text-white tracking-tight drop-shadow-sm">
        Orbital Path
      </h1>
      <p class="text-slate-600 dark:text-slate-400 text-xs md:text-sm font-medium mt-2">
         The nucleus of your Go knowledge
      </p>
    </div>

    <div v-if="loading" class="flex-grow flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold animate-pulse text-xl z-10">
      Calibrating Orbit...
    </div>


    <!-- Geometric Container (Forces square ratio for perfect circle) -->
    <div v-else class="flex-grow w-full flex items-center justify-center z-10 my-auto p-4">
      <!-- Actual Square Grid -->
      <div class="relative w-full" style="max-width: min(100%, 70vh, 800px); aspect-ratio: 1/1;">
        
        <!-- Orbital Paths SVG -->
        <svg viewBox="0 0 100 100" class="absolute inset-0 w-full h-full pointer-events-none drop-shadow-sm overflow-visible">
        
        <!-- Primary Electron Orbit (Perfect Circle mapping to radius = 40) -->
        <circle cx="50" cy="50" r="40" class="stroke-indigo-400/30 dark:stroke-indigo-400/30" stroke-width="0.2" stroke-dasharray="1 2" fill="none" />
        
        <!-- Nucleus (Center of the graph) -->
        <circle cx="50" cy="50" r="2.5" class="fill-indigo-500/40 dark:fill-indigo-400/40 animate-pulse" />
        <circle cx="50" cy="50" r="6" class="stroke-indigo-400/20 dark:stroke-indigo-300/20" stroke-width="0.1" fill="none" />

        <!-- Sequential progress connecting lines -->
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

        <!-- Closing loop line between last and first stage -->
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

        <!-- Nodes Array -->
        <div 
          v-for="(stage, index) in layoutNodes" 
          :key="stage.id"
          class="absolute transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer"
          :style="{ top: stage.topPercent, left: stage.leftPercent }"
          @click="startStage(stage)"
        >
          <div class="relative flex flex-col items-center justify-center transition-transform duration-300 group-hover:scale-110">
             
             <div class="relative flex items-center justify-center">
               <div v-if="getStageStatus(stage) !== 'locked'" 
                    class="absolute inset-0 rounded-full animate-ping opacity-20"
                    :class="getStageStatus(stage) === 'completed' ? 'bg-emerald-400' : 'bg-indigo-400'">
               </div>
               
               <div 
                 class="relative z-10 w-10 h-10 md:w-12 md:h-12 rounded-[14px] rotate-45 flex items-center justify-center shadow-lg border-2 transition-all duration-300 backdrop-blur-md"
                 :class="[
                    getStageStatus(stage) === 'locked' ? 'bg-slate-200/80 border-slate-300 dark:bg-slate-800/80 dark:border-slate-700 text-slate-400 dark:text-slate-600' :
                    getStageStatus(stage) === 'completed' ? 'bg-emerald-100 border-emerald-400 text-emerald-500 dark:bg-emerald-900/60 dark:border-emerald-500 shadow-[0_0_15px_rgba(16,185,129,0.4)]' :
                    'bg-indigo-100 border-indigo-500 text-indigo-600 dark:bg-indigo-900/60 dark:text-indigo-400 dark:border-indigo-400 shadow-[0_0_15px_rgba(99,102,241,0.5)]'
                 ]"
               >
                  <div class="-rotate-45 flex items-center justify-center">
                    <svg v-if="getStageStatus(stage) === 'locked'" class="w-4 h-4 md:w-5 md:h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" fill-rule="evenodd"></path></svg>
                    <svg v-else class="w-5 h-5 md:w-6 md:h-6" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
                  </div>
               </div>
             </div>

             <!-- Stage Title Pop-up on the RADIAL OUTER side of Node on Hover (Non-overlapping, Compact & Truncated) -->
             <div 
               class="absolute w-max max-w-[130px] sm:max-w-[160px] pointer-events-none opacity-0 group-hover:opacity-100 scale-90 group-hover:scale-100 transition-all duration-200 ease-out z-50"
               :class="stage.tooltipPos"
             >
               <div 
                 class="px-2.5 py-1 text-[11px] font-extrabold tracking-tight rounded-xl bg-slate-950/95 text-slate-100 border border-slate-700/80 shadow-2xl backdrop-blur-md text-center truncate"
                 :class="getStageStatus(stage) === 'locked' ? 'text-slate-400 border-slate-800' : 'text-indigo-200 border-indigo-500/50 shadow-[0_0_15px_rgba(99,102,241,0.25)]'"
               >
                 {{ stage.title }}
               </div>
             </div>
          </div>
        </div>

    </div>
  </div>
</div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  @apply bg-slate-300 dark:bg-slate-600 rounded-full;
}
</style>
