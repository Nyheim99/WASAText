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

		router.push("/");
	} catch (error) {
		console.error("Login failed:", error.response?.data || error.message);
		alert("Login failed. Please try again.");
	}
};
</script>

<template>
	<div class="d-flex justify-content-center align-items-center vh-100">
		<div class="card p-4" style="max-width: 300px">
			<h1 class="text-center mb-3">Login</h1>

			<input
				v-model="username"
				type="text"
				placeholder="Enter your username"
				class="form-control mb-3"
			/>

			<button
				class="btn btn-primary w-100"
				@click="handleLogin"
			>
				Login
			</button>
		</div>
	</div>
</template>
