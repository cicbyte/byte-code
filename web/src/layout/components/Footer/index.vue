<template>
  <div class="page-footer">
    <div class="copyright">ByteCode</div>
  </div>
</template>

<script>
  export default {
    name: 'PageFooter',
    components: {},
    props: {
      collapsed: {
        type: Boolean,
      },
    },
  };
</script>

<style lang="less" scoped>
  .page-footer {
    padding: 0 16px;
    text-align: center;

    .copyright {
      color: #808695;
      font-size: 14px;
    }
  }
</style>
