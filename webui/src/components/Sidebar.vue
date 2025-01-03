<script>
export default {
  props: {
    logout: Function,
  },
};
</script>

<template>
	<div class="sidebar">
		<button @click="logout" class="logout-button">Logout</button>
	</div>
</template>

<style>
.sidebar {
	background-color: #2c3e50;
	color: white;
	display: flex;
	flex-direction: column;
	justify-content: flex-end;
	padding: 20px;
}

.logout-button {
	background-color: #e74c3c;
	color: white;
	border: none;
	padding: 10px;
	font-size: 16px;
	cursor: pointer;
	border-radius: 5px;
}

.logout-button:hover {
  background-color: #c0392b;
}

</style>
