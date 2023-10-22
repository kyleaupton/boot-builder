import { createApp } from 'vue';
import { createPinia } from 'pinia';

// Icons
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
// Solid
import {
  faChevronUp,
  faCircleCheck,
  faCompactDisc,
  faDesktop,
  faDownLong,
  faFloppyDisk,
  faHardDrive,
  faTriangleExclamation,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
// Regular
import { faHardDrive as faRegHardDrive } from '@fortawesome/free-regular-svg-icons';
// Brands
import {
  faApple,
  faLinux,
  faUsb,
  faWindows,
} from '@fortawesome/free-brands-svg-icons';

import App from './App.vue';

library.add(
  // Solid
  faChevronUp,
  faCircleCheck,
  faCompactDisc,
  faDesktop,
  faDownLong,
  faFloppyDisk,
  faHardDrive,
  faTriangleExclamation,
  faXmark,
  // Regular
  faRegHardDrive,
  // Brands
  faApple,
  faLinux,
  faUsb,
  faWindows,
);

createApp(App)
  .component('font-awesome-icon', FontAwesomeIcon)
  .use(createPinia())
  .mount('#app')
  .$nextTick(() => {
    postMessage({ payload: 'removeLoading' }, '*');
  });
