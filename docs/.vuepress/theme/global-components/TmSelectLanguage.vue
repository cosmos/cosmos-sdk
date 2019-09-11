<template lang="pug">
  div
    .select
      select(v-model="language" @input="select")
        option(v-for="(value, key) in $site.themeConfig.locales" :value="key") {{value.label}}
</template>

<style lang="stylus" scoped>
select
  border none
  background none
  text-transform uppercase
  letter-spacing 0.02em
  font-weight 500
  font-size 0.875rem

.select
  border 2px solid rgba(140, 145, 177, 0.32)
  padding 0.25rem 0.5rem
  border-radius 6px
  background-color transparent
</style>

<script>
export default {
  data: function() {
    return {
      language: null
    };
  },
  created() {
    this.language = this.$localeConfig.path;
  },
  methods: {
    select(e) {
      this.$router.push(
        this.$page.path.replace(this.$localeConfig.path, e.target.value)
      );
    }
  }
};
</script>