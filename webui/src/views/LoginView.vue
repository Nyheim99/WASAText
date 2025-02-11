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
		<div
			class="p-4 shadow-sm"
			style="
				max-width: 300px;
				background-color: #C8E1FF;
				border-radius: 12px;
			"
		>
			<h1 class="text-center mb-3">Login</h1>

			<input
				type="text"
				v-model="username"
				placeholder="Enter your username"
				class="form-control mb-3"
				style="border: 1px solid #A7C7E7; background-color: #ffffff"
			/>

			<button
				@click="handleLogin"
				class="btn w-100"
				style="
					background-color: #4a90e2;
					color: white;
					border: none;
					transition: background-color 0.2s ease;
				"
				onmouseover="this.style.backgroundColor='#357ABD';"
				onmouseout="this.style.backgroundColor='#4A90E2';"
			>
				Login
			</button>
		</div>
	</div>
</template>
