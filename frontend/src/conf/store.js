import { createStore } from 'vuex';

const store = createStore({
  state: {
    busy: false,
    port: 0
  }
});

export default store;