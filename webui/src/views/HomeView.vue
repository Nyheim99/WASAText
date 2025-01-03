<script>
import { useRouter } from "vue-router";

export default {
  setup() {
    const router = useRouter();

    const logout = () => {
      // Remove userId from localStorage
      localStorage.removeItem("userId");

      // Redirect to login page
      router.push("/login");
    };

    return {
      logout,
    };
  },
};

</script>

<template>
	<div>
		<h1>Welcome to the Home Page!</h1>
		<p>You are logged in.</p>
		<!-- Logout Button -->
		<button @click="logout">Logout</button>
	</div>
</template>

<style>
/* Add some styling to the logout button */
button {
  background-color: #ff4d4d;
  color: white;
  border: none;
  padding: 10px 20px;
  font-size: 16px;
  cursor: pointer;
  border-radius: 5px;
  margin-top: 20px;
}

button:hover {
  background-color: #ff1a1a;
}
</style>
