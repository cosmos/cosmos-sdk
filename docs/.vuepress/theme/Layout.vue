<template lang="pug">
  div
    tm-layout
      template(v-slot:content)
        Content
      template(v-slot:sidebar)
        tm-content-sidebar
</template>