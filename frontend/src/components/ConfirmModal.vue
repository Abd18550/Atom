<script setup>
defineProps({
  show: Boolean,
  title: String,
  message: String,
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' }
})
defineEmits(['confirm', 'cancel'])
</script>
<template>
  <div v-if="show" class="fixed inset-0 z-[100] overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity bg-slate-900/60 backdrop-blur-sm" aria-hidden="true" @click="$emit('cancel')"></div>
      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div class="inline-block overflow-hidden text-left align-bottom transition-all transform bg-white dark:bg-slate-800 rounded-2xl shadow-2xl sm:my-8 sm:align-middle w-full max-w-md border border-slate-200 dark:border-slate-700">
        <div class="px-4 pt-5 pb-4 bg-white dark:bg-slate-800 sm:p-6 sm:pb-4">
          <div class="sm:flex sm:items-start">
            <div class="flex items-center justify-center flex-shrink-0 w-12 h-12 mx-auto bg-red-100 dark:bg-red-900/30 rounded-full sm:mx-0 sm:h-10 sm:w-10">
              <svg class="w-6 h-6 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
              <h3 class="text-lg font-bold leading-6 text-slate-900 dark:text-white" id="modal-title">{{ title }}</h3>
              <div class="mt-2">
                <p class="text-sm font-medium text-slate-500 dark:text-slate-400">{{ message }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="px-4 py-3 bg-slate-50 dark:bg-slate-900/50 sm:flex sm:flex-row-reverse sm:px-6 border-t border-slate-100 dark:border-slate-700/50">
          <button type="button" class="inline-flex justify-center w-full px-4 py-2 text-sm font-bold text-white bg-red-600 rounded-lg shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto transition-colors focus:ring-2 focus:ring-red-500 focus:outline-none" @click="$emit('confirm')">{{ confirmText }}</button>
          <button type="button" class="inline-flex justify-center w-full px-4 py-2 mt-3 text-sm font-bold text-slate-700 bg-white dark:bg-slate-800 dark:text-slate-200 rounded-lg shadow-sm border border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700 sm:mt-0 sm:w-auto transition-colors focus:outline-none" @click="$emit('cancel')">{{ cancelText }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
