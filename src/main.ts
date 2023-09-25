import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import {
  faFloppyDisk,
  faHardDrive,
  faCompactDisc,
} from '@fortawesome/free-solid-svg-icons';
import { faUsb } from '@fortawesome/free-brands-svg-icons';
import App from './App.vue';

library.add(faUsb, faFloppyDisk, faHardDrive, faCompactDisc);

createApp(App)
  .component('font-awesome-icon', FontAwesomeIcon)
  .use(createPinia())
  .mount('#app')
  .$nextTick(() => {
    postMessage({ payload: 'removeLoading' }, '*');
  });
