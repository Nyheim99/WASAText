<script setup>
import { ref } from "vue";
import axios from "../services/axios";
import { useRouter } from "vue-router";

const username = ref("");
const router = useRouter();

const handleLogin = async () => {
	try {
		const response = await axios.post("/session", {
			username: username.value,
		});
		const userId = response.data.identifier;

		localStorage.setItem("userId", userId);

		console.log("Login successful:", response.data);

    const usersResponse = await axios.get("/users");
    console.log("Fetched users:", usersResponse.data);

    const conversationsResponse = await axios.get("/user/conversations")
    console.log("Fetched conversations:", conversationsResponse.data);

		router.push("/");
	} catch (error) {
		console.error("Login failed:", error.response?.data || error.message);
		alert("Login failed. Please try again.");
	}
};
</script>

<template>
	<div class="login-wrapper">
		<div id="login-page" class="login-container">
			<h1>Login</h1>

			<input
				type="text"
				v-model="username"
				placeholder="Enter your username"
				class="username-input"
			/>

			<button @click="handleLogin" class="login-button">Login</button>
		</div>
	</div>
</template>

<style>
body,
html {
	margin: 0;
	padding: 0;
	height: 100%;
	font-family: Arial, sans-serif;
}

.login-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background-color: #f5f5f5;
}

.login-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
	background-color: white;
	padding: 20px;
	box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	border-radius: 10px;
	max-width: 300px;
	margin: auto;
}

h1 {
	font-size: 2rem;
	margin-bottom: 10px;
	color: #333;
}

.username-input {
	width: 100%;
	padding: 10px;
	margin-bottom: 15px;
	border: 1px solid #ccc;
	border-radius: 5px;
	font-size: 1rem;
}

.login-button {
	padding: 10px 20px;
	background-color: #007bff;
	color: white;
	border: none;
	border-radius: 5px;
	font-size: 1rem;
	cursor: pointer;
	transition: background-color 0.3s;
}

.login-button:hover {
	background-color: #0056b3;
}
</style>
