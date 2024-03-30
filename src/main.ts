import { createApp } from 'vue';
import { createPinia } from 'pinia';

import PrimeVue from 'primevue/config';
// Components
import Button from 'primevue/button';
import Dropdown from 'primevue/dropdown';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import InputText from 'primevue/inputtext';
import ProgressBar from 'primevue/progressbar';
import 'primevue/resources/themes/lara-dark-amber/theme.css';
import 'primeicons/primeicons.css';

// Icons
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
// Solid
import {
  faBan,
  faCircleCheck,
  faDisplay,
  faFile,
  faHardDrive,
} from '@fortawesome/free-solid-svg-icons';
// Regular
// Brands
import { faApple, faWindows } from '@fortawesome/free-brands-svg-icons';

import App from './App.vue';

library.add(
  // Solid
  faBan,
  faCircleCheck,
  faDisplay,
  faFile,
  faHardDrive,
  // Regular
  // Brands
  faApple,
  faWindows,
);

createApp(App)
  // Font Awesome
  .component('font-awesome-icon', FontAwesomeIcon)
  // PrimeVue
  .component('Button', Button)
  .component('Dropdown', Dropdown)
  .component('InputGroup', InputGroup)
  .component('InputGroupAddon', InputGroupAddon)
  .component('InputText', InputText)
  .component('ProgressBar', ProgressBar)
  .use(createPinia())
  .use(PrimeVue, { inputStyle: 'filled' })
  .mount('#app')
  .$nextTick(() => {
    postMessage({ payload: 'removeLoading' }, '*');
  });
