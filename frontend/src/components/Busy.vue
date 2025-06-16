<template>
  <div class="busy-overlay">
    <div class="busy-spinner"></div>
    <div v-if="progress !== null" class="busy-text">Processing: {{ progress.toFixed(2) }}%</div>
    <div v-else class="busy-text">Processing...</div>
  </div>
</template>

<script>
export default {
  name: 'Busy',
  data() {
    return {
      progress: null
    };
  },
  mounted() {
    this.$bus.on('progress', (progress) => {
      this.progress = progress;
    });
  },
  beforeDestroy() {
    this.$bus.off('progress');
  },
};
</script>

<style scoped>
.busy-overlay {
  position: absolute;
  z-index: 1000;
  top: 0;
  left: 0;
  height: 100vh;
  width: 100vw;
  background: rgba(30, 30, 30, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  pointer-events: all;
  user-select: none;
}

.busy-spinner {
  border: 6px solid #eee;
  border-top: 6px solid #1976d2;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

.busy-text {
  color: #fff;
  font-size: 1.2em;
  font-weight: bold;
  text-align: center;
  font-family: 'Fira Mono', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', monospace;
}
</style>