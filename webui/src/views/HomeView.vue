<script>
import { useRouter } from "vue-router";
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import Sidebar from "../components/Sidebar.vue";
import ConversationList from "../components/ConversationList.vue";

export default {
	components: {
		Sidebar,
		ConversationList,
	},
	setup() {
		const router = useRouter();
		const conversations = ref([]);
		const user = ref({});
		const feedbackMessage = ref("");
		const showFeedback = ref(false);

		const logout = () => {
			localStorage.removeItem("userId");
			router.push("/login");
		};

		const updateUsername = async (newUsername) => {
			try {
				const response = await axios.put("/user/username", {
					username: newUsername,
				});
				console.log("Username updated successfully:", response.data);
				feedbackMessage.value = "Username updated successfully!";
				showFeedback.value = true;

				user.value.username = newUsername;

				setTimeout(() => {
					showFeedback.value = false;
				}, 3000);

				return true;
			} catch (error) {
				console.error(
					"Failed to update username:",
					error.response?.data || error.message
				);
				return error.response?.status === 409
					? "Username is already in use."
					: "Invalid username or another error occurred.";
			}
		};

		const fetchUser = async () => {
			try {
				const response = await axios.get("/user");
				user.value = response.data;
			} catch (error) {
				console.error("Failed to fetch user:", error);
			} 
		};

		const fetchConversations = async () => {
			try {
				const response = await axios.get("/user/conversations");
				conversations.value = response.data || [];
			} catch (error) {
				console.error("Failed to fetch conversations:", error.message);
				conversations.value = [];
			}
		};

		onMounted(() => {
			fetchUser();
			fetchConversations();
		});

		return {
			logout,
			updateUsername,
			user,
			conversations,
			feedbackMessage,
			showFeedback,
		};
	},
};
</script>

<template>
	<div class="container-fluid d-flex vh-100 flex-column p-3">
		<div
			v-if="showFeedback"
			class="alert alert-success position-fixed top-2 start-50 translate-middle-x shadow"
			role="alert"
			style="z-index: 1050"
		>
			{{ feedbackMessage }}
		</div>

		<div class="row flex-grow-1 g-3">
			<div class="col-auto p-0">
				<Sidebar
					:logout="logout"
					:user="user"
					:updateUsername="updateUsername"
				/>
			</div>
			<div class="col-3">
				<ConversationList :conversations="conversations" />
			</div>
			<div class="col">
				<div class="bg-white shadow-sm rounded p-4 overflow-auto h-100">
					<h2>Chat Window</h2>
					<p>Select a chat to start messaging!</p>
				</div>
			</div>
		</div>
	</div>
</template>
