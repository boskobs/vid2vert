<template>
  <div class="VideoPlayer">
    <div class="VideoPlayerContainer">
      <video ref="videoRef" :controls="!resizingState" :src="src" @error="$emit('error')"
        @loadedmetadata="onLoadedMetadata" @timeupdate="onTimeUpdate">
      </video>
      <div v-if="videoWidth && videoHeight" class="rect-overlay" :style="rectStyle" @mousedown="startDrag($event)"
        @touchstart="startDrag($event, true)">
        <div v-if="resizingState" class="resize-handle" @mousedown.stop="startResize($event)" @touchstart.stop="startResize($event, true)">
        </div>
        <div v-if="!resizingState" class="arrow arrow-x" @mousedown.stop="startDragX($event)"
          @touchstart.stop="startDragX($event, true)" :title="'Drag horizontally'">
          <svg width="32" height="24" viewBox="0 0 32 24">
            <line x1="4" y1="12" x2="28" y2="12" stroke="#42b983" stroke-width="4" stroke-linecap="round" />
            <polygon points="32,12 22,8 22,16" fill="#42b983" />
            <polygon points="0,12 10,8 10,16" fill="#42b983" />
          </svg>
        </div>
        <div v-if="!resizingState" class="arrow arrow-y" @mousedown.stop="startDragY($event)"
          @touchstart.stop="startDragY($event, true)" :title="'Drag vertically'">
          <svg width="24" height="32" viewBox="0 0 24 32">
            <line x1="12" y1="4" x2="12" y2="28" stroke="#42b983" stroke-width="4" stroke-linecap="round" />
            <polygon points="12,32 8,22 16,22" fill="#42b983" />
            <polygon points="12,0 8,10 16,10" fill="#42b983" />
          </svg>
        </div>
        <button v-if="resizingState" class="confirm-btn" @click="resizingState = false">Confirm</button>
      </div>
    </div>
    <Timeline v-if="!resizingState" :keyframes="keyframeList" :duration="$refs.videoRef?.duration || 0"
      :currentTime="currentTime" @seek="seekVideo" @delete="deleteKeyframe($event)" />
  </div>
</template>

<script>
import Timeline from './Timeline.vue';

export default {
  name: 'VideoPlayer',
  components: { Timeline },
  props: {
    src: { type: String, required: true },
    aspectW: { type: Number, required: true },
    aspectH: { type: Number, required: true },
  },
  emits: ['error'],
  data() {
    return {
      resizingState: true,
      videoWidth: 0,
      videoHeight: 0,
      rect: {
        x: 0, // percent
        y: 0, // percent
        w: 30, // percent
        h: 30, // percent
      },
      dragging: false,
      resizing: false,
      dragAxis: null,
      dragOffset: { x: 0, y: 0 },
      keyframes: {},
      currentTime: 0,
    };
  },
  computed: {
    aspectRatio() {
      return this.aspectW / this.aspectH;
    },
    keyframeList() {
      // Convert keyframes object to sorted array
      return Object.values(this.keyframes).sort((a, b) => a.time - b.time);
    },
    interpolatedRect() {
      if (this.dragging || this.resizing || this.keyframeList.length === 0) return this.rect;
      const t = this.currentTime;
      const frames = this.keyframeList;
      if (t <= frames[0].time) return frames[0];
      if (t >= frames[frames.length - 1].time) return frames[frames.length - 1];
      // Find the two keyframes surrounding t
      let i = 0;
      while (i < frames.length - 1 && frames[i + 1].time <= t) i++;
      const kf1 = frames[i];
      const kf2 = frames[i + 1];
      const dt = kf2.time - kf1.time;
      const alpha = dt === 0 ? 0 : (t - kf1.time) / dt;
      // Linear interpolation
      return {
        x: kf1.x + (kf2.x - kf1.x) * alpha,
        y: kf1.y + (kf2.y - kf1.y) * alpha,
        w: kf1.w + (kf2.w - kf1.w) * alpha,
        h: kf1.h + (kf2.h - kf1.h) * alpha,
      };
    },
    rectStyle() {
      // Use interpolated rect unless dragging/resizing
      const r = this.interpolatedRect;
      const left = (r.x / 100) * this.videoWidth;
      const top = (r.y / 100) * this.videoHeight;
      const width = (r.w / 100) * this.videoWidth;
      const height = (r.h / 100) * this.videoHeight;
      return {
        left: left + 'px',
        top: top + 'px',
        width: width + 'px',
        height: height + 'px',
        cursor: this.dragging ? 'grabbing' : 'grab'
      };
    },
  },
  mounted() {
    this.initResizeObserver();
  },
  beforeUnmount() {
    this.stopDrag();
    this.stopResize();
    if (this.resizeObserver) this.resizeObserver.disconnect();
  },
  methods: {
    initResizeObserver() {
      const video = this.$refs.videoRef;
      if (!video) return;
      this.resizeObserver = new ResizeObserver(entries => {
        for (let entry of entries) {
          this.videoWidth = entry.contentRect.width;
          this.videoHeight = entry.contentRect.height;
        }
      });
      this.resizeObserver.observe(video);
    },
    async onLoadedMetadata(e) {
      const video = this.$refs.videoRef;
      this.videoWidth = video.clientWidth;
      this.videoHeight = video.clientHeight;
      this.initResizeObserver();
      this.setRectToAspect();
      this.addKeyframe();
    },
    setRectToAspect() {
      // Set initial rect size to fit aspect ratio and be centered
      let vw = this.videoWidth, vh = this.videoHeight;
      let rw = vw * 0.5, rh = rw / this.aspectRatio;
      if (rh > vh * 0.5) {
        rh = vh * 0.5;
        rw = rh * this.aspectRatio;
      }
      this.rect.w = (rw / vw) * 100;
      this.rect.h = (rh / vh) * 100;
      this.rect.x = 0;
      this.rect.y = 0;
    },
    startDrag(e, isTouch = false) {
      if (this.resizingState) return;
      this.dragging = true;
      this.dragAxis = null;
      const evt = isTouch ? e.touches[0] : e;
      const overlay = e.currentTarget;
      const overlayRect = overlay.getBoundingClientRect();
      this.dragOffset = {
        x: evt.clientX - overlayRect.left,
        y: evt.clientY - overlayRect.top,
      };
      window.addEventListener(isTouch ? 'touchmove' : 'mousemove', this.onDragMove, { passive: false });
      window.addEventListener(isTouch ? 'touchend' : 'mouseup', this.stopDrag, { passive: false });
    },
    startDragX(e, isTouch = false) {
      this.dragging = true;
      this.dragAxis = 'x';
      const evt = isTouch ? e.touches[0] : e;
      const overlay = e.currentTarget.parentElement;
      const overlayRect = overlay.getBoundingClientRect();
      this.dragOffset = {
        x: evt.clientX - overlayRect.left,
        y: evt.clientY - overlayRect.top,
      };
      window.addEventListener(isTouch ? 'touchmove' : 'mousemove', this.onDragMove, { passive: false });
      window.addEventListener(isTouch ? 'touchend' : 'mouseup', this.stopDrag, { passive: false });
    },
    startDragY(e, isTouch = false) {
      this.dragging = true;
      this.dragAxis = 'y';
      const evt = isTouch ? e.touches[0] : e;
      const overlay = e.currentTarget.parentElement;
      const overlayRect = overlay.getBoundingClientRect();
      this.dragOffset = {
        x: evt.clientX - overlayRect.left,
        y: evt.clientY - overlayRect.top,
      };
      window.addEventListener(isTouch ? 'touchmove' : 'mousemove', this.onDragMove, { passive: false });
      window.addEventListener(isTouch ? 'touchend' : 'mouseup', this.stopDrag, { passive: false });
    },
    onDragMove(e) {
      if (!this.dragging) return;
      const evt = e.touches ? e.touches[0] : e;
      const wrapperRect = this.$refs.videoRef.getBoundingClientRect();
      let x = evt.clientX - wrapperRect.left - this.dragOffset.x;
      let y = evt.clientY - wrapperRect.top - this.dragOffset.y;
      const rectPx = this.getRectPx();
      // Clamp within video
      x = Math.max(0, Math.min(x, this.videoWidth - rectPx.width));
      y = Math.max(0, Math.min(y, this.videoHeight - rectPx.height));
      if (this.dragAxis === 'x') {
        this.rect.x = (x / this.videoWidth) * 100;
      } else if (this.dragAxis === 'y') {
        this.rect.y = (y / this.videoHeight) * 100;
      } else {
        this.rect.x = (x / this.videoWidth) * 100;
        this.rect.y = (y / this.videoHeight) * 100;
      }
    },
    stopDrag(e) {
      this.dragging = false;
      this.dragAxis = null;
      window.removeEventListener('mousemove', this.onDragMove);
      window.removeEventListener('mouseup', this.stopDrag);
      window.removeEventListener('touchmove', this.onDragMove);
      window.removeEventListener('touchend', this.stopDrag);
      this.addKeyframe();
    },
    startResize(e, isTouch = false) {
      if (this.dragging) return;
      this.resizing = true;
      this.resizeStart = isTouch ? e.touches[0] : e;
      this.startRect = { ...this.rect };
      window.addEventListener(isTouch ? 'touchmove' : 'mousemove', this.onResizeMove, { passive: false });
      window.addEventListener(isTouch ? 'touchend' : 'mouseup', this.stopResize, { passive: false });
    },
    onResizeMove(e) {
      if (!this.resizing) return;
      const evt = e.touches ? e.touches[0] : e;
      const videoRect = this.$refs.videoRef.getBoundingClientRect();
      let dx = evt.clientX - this.resizeStart.clientX;
      let dy = evt.clientY - this.resizeStart.clientY;
      // Resize keeping aspect ratio
      let dpx = dx;
      // Use horizontal movement for resizing
      let vw = this.videoWidth, vh = this.videoHeight;
      let pxW = (this.startRect.w / 100) * vw + dpx;
      let pxH = pxW / this.aspectRatio;
      // Clamp to video bounds
      if (pxW < 20) { pxW = 20; pxH = pxW / this.aspectRatio; }
      if (pxH < 20) { pxH = 20; pxW = pxH * this.aspectRatio; }
      if (pxW + (this.startRect.x / 100) * vw > vw) {
        pxW = vw - (this.startRect.x / 100) * vw;
        pxH = pxW / this.aspectRatio;
      }
      if (pxH + (this.startRect.y / 100) * vh > vh) {
        pxH = vh - (this.startRect.y / 100) * vh;
        pxW = pxH * this.aspectRatio;
      }
      this.rect.w = (pxW / vw) * 100;
      this.rect.h = (pxH / vh) * 100;
    },
    stopResize(e) {
      this.resizing = false;
      window.removeEventListener('mousemove', this.onResizeMove);
      window.removeEventListener('mouseup', this.stopResize);
      window.removeEventListener('touchmove', this.onResizeMove);
      window.removeEventListener('touchend', this.stopResize);
      this.addKeyframe();
    },
    addKeyframe() {
      const video = this.$refs.videoRef;
      if (!video) return;
      this.keyframes[video.currentTime] = {
        time: video.currentTime,
        x: this.rect.x,
        y: this.rect.y,
        w: this.rect.w,
        h: this.rect.h
      };
      console.log(this.keyframes[video.currentTime]);
    },
    deleteKeyframe(time) {
      delete this.keyframes[time];
    },
    onTimeUpdate(e) {
      this.currentTime = this.$refs.videoRef.currentTime;
    },
    seekVideo(time) {
      if (this.$refs.videoRef) {
        this.$refs.videoRef.currentTime = time;
      }
    },
    getRectPx() {
      return {
        left: (this.rect.x / 100) * this.videoWidth,
        top: (this.rect.y / 100) * this.videoHeight,
        width: (this.rect.w / 100) * this.videoWidth,
        height: (this.rect.h / 100) * this.videoHeight,
      };
    }
  }
};
</script>

<style scoped>
.VideoPlayer {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.VideoPlayerContainer {
  position: relative;
  width: 80%;
  user-select: none;
}

video {
  display: block;
  width: 100%;
  overflow: hidden;
  height: auto;
  z-index: 1;
}

.rect-overlay {
  position: absolute;
  pointer-events: auto;
  box-sizing: border-box;
  border: 2px solid #42b983;
  box-sizing: border-box;
  z-index: 2;
  background: rgb(66, 185, 131, 0.08);
}

.arrow {
  position: absolute;
  z-index: 3;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.8;
  transition: opacity 0.2s, filter 0.2s;
}

.arrow:hover {
  opacity: 1;
  filter: drop-shadow(0 0 6px #42b983);
}

.arrow-x {
  left: 50%;
  top: 100%;
  transform: translate(-50%, 0);
}

.arrow-y {
  left: 100%;
  top: 50%;
  transform: translate(0, -50%);
}

.arrow svg {
  pointer-events: none;
}

.resize-handle {
  position: absolute;
  right: -8px;
  bottom: -8px;
  width: 16px;
  height: 16px;
  background: #42b983;
  border-radius: 50%;
  border: 2px solid #fff;
  cursor: nwse-resize;
  z-index: 3;
}

.confirm-btn {
  position: absolute;
  right: -100px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 10;
  background: #42b983;
  color: #fff;
  border: none;
  border-radius: 4px;
  padding: 6px 14px;
  font-size: 1em;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0,0,0,0.12);
  transition: background 0.2s;
}
.confirm-btn:hover {
  background: #36996b;
}

:deep(.timeline) {
  width: 80%;
  margin-left: auto;
  margin-right: auto;
}
</style>
