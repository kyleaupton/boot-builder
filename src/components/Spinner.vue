<template>
  <div class="wrapper">
    <div class="wrapper2">
      <div class="spinner">
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
        <div class="spinner-blade"></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Spinner',
});
</script>

<style scoped>
.wrapper {
  width: 32px;
  height: 32px;
  position: relative;
}

.wrapper2 {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.spinner {
  height: 100%;
  width: 100%;
  font-size: 75px;
  position: relative;
}

.spinner-blade {
  /* This kinda makes it centered...? Really janky tho */
  position: absolute;
  top: 65%;
  left: 48%;

  width: 2px;
  height: 6px;
  border-radius: 0.5em;
  background-color: transparent;
  transform-origin: center -4px;
  animation: spinner-fade 1s infinite linear;
}

.spinner-blade:nth-child(1) {
  -webkit-animation-delay: 0s;
  animation-delay: 0s;
  transform: rotate(0deg);
}
.spinner-blade:nth-child(2) {
  -webkit-animation-delay: 0.083s;
  animation-delay: 0.083s;
  transform: rotate(30deg);
}
.spinner-blade:nth-child(3) {
  -webkit-animation-delay: 0.166s;
  animation-delay: 0.166s;
  transform: rotate(60deg);
}
.spinner-blade:nth-child(4) {
  -webkit-animation-delay: 0.249s;
  animation-delay: 0.249s;
  transform: rotate(90deg);
}
.spinner-blade:nth-child(5) {
  -webkit-animation-delay: 0.332s;
  animation-delay: 0.332s;
  transform: rotate(120deg);
}
.spinner-blade:nth-child(6) {
  -webkit-animation-delay: 0.415s;
  animation-delay: 0.415s;
  transform: rotate(150deg);
}
.spinner-blade:nth-child(7) {
  -webkit-animation-delay: 0.498s;
  animation-delay: 0.498s;
  transform: rotate(180deg);
}
.spinner-blade:nth-child(8) {
  -webkit-animation-delay: 0.581s;
  animation-delay: 0.581s;
  transform: rotate(210deg);
}
.spinner-blade:nth-child(9) {
  -webkit-animation-delay: 0.664s;
  animation-delay: 0.664s;
  transform: rotate(240deg);
}
.spinner-blade:nth-child(10) {
  -webkit-animation-delay: 0.747s;
  animation-delay: 0.747s;
  transform: rotate(270deg);
}
.spinner-blade:nth-child(11) {
  -webkit-animation-delay: 0.83s;
  animation-delay: 0.83s;
  transform: rotate(300deg);
}
.spinner-blade:nth-child(12) {
  -webkit-animation-delay: 0.913s;
  animation-delay: 0.913s;
  transform: rotate(330deg);
}

@-webkit-keyframes spinner-fade {
  0% {
    background-color: #69717d;
  }
  100% {
    background-color: transparent;
  }
}

@keyframes spinner-fade {
  0% {
    background-color: #69717d;
  }
  100% {
    background-color: transparent;
  }
}
</style>
