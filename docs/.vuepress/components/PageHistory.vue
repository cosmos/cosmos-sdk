<template>
  <div>
    {{url}}
  </div>
</template>

<script>
  export default {
    computed: {
      url() {
        const c = this.$themeConfig
        if (c && c.repo && c.docsDir && c.docsBranch) {
          return `https://github.com/${c.repo}/tree/${c.docsBranch}/${c.docsDir}/${this.$page.relativePath}`
        }
      },
    },
  }
</script>
