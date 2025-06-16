<template>
  <main>
    <Busy v-if="$store.state.busy" />
    <span>Aspect ratio: <input :disabled="!!video" v-model.number="aspectW" type="number"> : <input :disabled="!!video"
        v-model.number="aspectH" type="number"></span>
    <template v-if="ready">
      <button v-if="!video" @click="pickVideo">Pick video</button>
      <template v-if="video">
        <p>Video name: {{ video.name }}</p>
        <VideoPlayer ref="refVideoPlayer" :src="`http://localhost:${$store.state.port}/lastVideo`" controls
          style="width: 100%; object-fit: cover;" @error="video = null" :aspectW="aspectW" :aspectH="aspectH" />
        <button @click="video = null">Clear video</button>
        <button @click="saveChanges">Save changes</button>
      </template>
    </template>
  </main>
</template>

<script>
import Busy from './components/Busy.vue';
import VideoPlayer from './components/VideoPlayer.vue';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { Quit, HasFFmpeg, GetMediaServerPort, OpenVideo, SaveVideo } from '../wailsjs/go/main/App';

export default {
  components: { Busy, VideoPlayer },
  data() {
    return {
      ready: false,
      showSettings: false,
      aspectW: 9,
      aspectH: 16,
      video: null
    };
  },
  async mounted() {
    this.$store.state.port = await GetMediaServerPort();
    if (!(await HasFFmpeg())) {
      await this.$alert('FFmpeg is not installed. Please install it to use this app.\n On windows run "winget install ffmpeg" in a terminal.')
      return await Quit();
    }
    EventsOn('app:progress', (progress) => this.$bus.emit('progress', progress));
    this.ready = true;
  },
  methods: {
    async pickVideo() {
      try {
        this.video = {
          ...(await OpenVideo()),
          url: `http://localhost:${this.$store.state.port}/lastVideo`,
        };
      } catch (e) {
        console.error('Error picking video:', e);
        this.$alert('Error picking video: ' + e.message);
      }
    },
    async saveChanges() {
      if (!this.$refs.refVideoPlayer.keyframeList.length) return this.$alert('No keyframes to save.');
      try {
        this.$store.state.busy = true;
        await SaveVideo(this.video.fullPath, this.$refs.refVideoPlayer.keyframeList);
        this.$store.state.busy = false;
        this.$alert('Video saved successfully!');
      } catch (e) {
        console.error('Error saving video:', e);
        this.$alert('Error saving video: ' + e.message);
      }
    },
  },
}
</script>

<style scoped>
main {
  padding: 20px;
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100vh;
}

main>span>input[type="number"] {
  width: 45px;
  margin: 0 10px;
}
</style>