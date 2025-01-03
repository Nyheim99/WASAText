<script setup>
import { RouterLink, RouterView } from "vue-router";
</script>
<script>
export default {};
</script>

<template>
	<main class="main-container">
		<RouterView />
	</main>
</template>

<style>
html,
body,
#app {
	background-color: lightgray;
	height: 100%;
}
.main-container {
	height: 100%;
	padding: 20px
}
</style>
