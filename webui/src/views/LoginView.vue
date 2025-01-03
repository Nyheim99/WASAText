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
  <div class="d-flex justify-content-center align-items-center vh-100 bg-transparent">
    <div class="card p-4 shadow-sm" style="max-width: 300px;">
      <h1 class="text-center mb-3">Login</h1>

      <input
        type="text"
        v-model="username"
        placeholder="Enter your username"
        class="form-control mb-3"
      />

      <button @click="handleLogin" class="btn btn-primary w-100">Login</button>
    </div>
  </div>
</template>