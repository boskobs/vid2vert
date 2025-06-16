<template>
  <div class="timeline" @click="seek($event)">
    <div class="bar"></div>
    <div
      v-for="(kf, idx) in keyframes"
      :key="idx"
      class="keyframe-group"
      :style="{ left: (kf.time / duration * 100) + '%' }"
    >
      <div
        class="keyframe"
        :title="'t=' + kf.time.toFixed(2) + 's'"
      ></div>
      <button v-if="kf !== keyframes[0]" class="delete-btn" @click.stop="$emit('delete', kf.time)">✕</button>
    </div>
    <div class="playhead" :style="{ left: (currentTime / duration * 100) + '%' }"></div>
  </div>
</template>

<script>
export default {
  name: 'Timeline',
  props: {
    keyframes: { type: Array, required: true },
    duration: { type: Number, required: true },
    currentTime: { type: Number, required: true },
  },
  emits: ['seek', 'delete'],
  methods: {
    seek(e) {
      const rect = e.currentTarget.getBoundingClientRect();
      const percent = (e.clientX - rect.left) / rect.width;
      this.$emit('seek', percent * this.duration);
    },
  },
};
</script>

<style scoped>
.timeline {
  position: relative;
  height: 24px;
  background: #222;
  margin: 12px 0;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  margin-bottom: 40px;
}
.bar {
  position: absolute;
  top: 50%;
  left: 0;
  width: 100%;
  height: 4px;
  background: #444;
  transform: translateY(-50%);
  border-radius: 2px;
}
.keyframe-group {
  position: absolute;
  top: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  transform: translateX(-50%);
  z-index: 2;
}
.keyframe {
  width: 8px;
  height: 24px;
  background: #42b983;
  border-radius: 2px;
  box-shadow: 0 0 2px #000a;
}
.delete-btn {
  margin-top: 12px;
  font-size: 12px;
  background: #e74c3c;
  color: #fff;
  border: none;
  border-radius: 2px;
  padding: 0 4px;
  cursor: pointer;
  outline: none;
  transition: background 0.2s;
}
.delete-btn:hover {
  background: #c0392b;
}
.playhead {
  position: absolute;
  top: 0;
  width: 2px;
  height: 100%;
  background: #fff;
  transform: translateX(-50%);
  z-index: 3;
}
</style>
