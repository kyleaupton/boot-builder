import { createApp } from 'vue';
import { createPinia } from 'pinia';
import PrimeVue from 'primevue/config';
import 'primevue/resources/themes/lara-dark-amber/theme.css';
import 'primeicons/primeicons.css';

// Icons
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
// Solid
import {
  faArrowUpFromBracket,
  faChevronUp,
  faCircleCheck,
  faCompactDisc,
  faDesktop,
  faDisplay,
  faDownLong,
  faFloppyDisk,
  faFile,
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
  faArrowUpFromBracket,
  faChevronUp,
  faCircleCheck,
  faCompactDisc,
  faDesktop,
  faDisplay,
  faDownLong,
  faFloppyDisk,
  faFile,
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
  .use(PrimeVue, { inputStyle: 'filled' })
  .mount('#app')
  .$nextTick(() => {
    postMessage({ payload: 'removeLoading' }, '*');
  });
