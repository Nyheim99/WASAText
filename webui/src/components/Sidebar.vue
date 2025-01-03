<script>
export default {
  props: {
    logout: Function,
  },
};
</script>

<template>
  <div class="h-100 rounded text-white d-flex flex-column justify-content-end p-3">
    <button @click="logout" class="btn btn-danger">Logout</button>
  </div>
</template>