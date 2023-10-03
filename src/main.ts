import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import {
  faFloppyDisk,
  faHardDrive,
  faCompactDisc,
  faXmark,
  faCircleCheck,
  faChevronUp,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { faHardDrive as faRegHardDrive } from '@fortawesome/free-regular-svg-icons';
import { faUsb, faWindows } from '@fortawesome/free-brands-svg-icons';
import App from './App.vue';

library.add(
  faUsb,
  faFloppyDisk,
  faHardDrive,
  faRegHardDrive,
  faCompactDisc,
  faXmark,
  faCircleCheck,
  faWindows,
  faChevronUp,
  faTriangleExclamation,
);

createApp(App)
  .component('font-awesome-icon', FontAwesomeIcon)
  .use(createPinia())
  .mount('#app')
  .$nextTick(() => {
    postMessage({ payload: 'removeLoading' }, '*');
  });
