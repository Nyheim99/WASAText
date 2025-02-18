<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "../services/axios";

const username = ref("");
const router = useRouter();

//Function to handle login
const handleLogin = async () => {
	//Attempt login
	try {
		const response = await axios.post("/session", {
			username: username.value,
		});
		const userId = response.data.identifier;

		//Set userID in localstorage
		localStorage.setItem("userId", userId);

		//Redirect to homepage on successful login
		router.push("/");
	} catch (error) {
		console.error("Login failed:", error.response?.data || error.message);
		alert("Login failed. Please try again.");
	}
};
</script>

<template>
	<div class="d-flex justify-content-center align-items-center vh-100">
		<!-- Login card-->
		<div class="card p-4" style="max-width: 300px">
			<h1 class="text-center mb-3">Login</h1>

			<input
				v-model="username"
				type="text"
				placeholder="Enter your username"
				class="form-control mb-3"
			/>

			<button class="btn btn-primary w-100" @click="handleLogin">
				Login
			</button>
		</div>
	</div>
</template>
