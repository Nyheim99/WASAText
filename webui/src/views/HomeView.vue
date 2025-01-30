<script>
import { useRouter } from "vue-router";
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import Sidebar from "../components/Sidebar.vue";
import ConversationList from "../components/ConversationList.vue";
import Conversation from "../components/Conversation.vue";

export default {
	components: {
		Sidebar,
		ConversationList,
		Conversation,
	},
	setup() {
		const router = useRouter();
		const user = ref({});
		const feedbackMessage = ref("");
		const showFeedback = ref(false);
		const conversations = ref([]);
		const selectedConversation = ref(null);
		const allUsers = ref([]);

		const fetchUser = async () => {
			try {
				const response = await axios.get("/user");
				user.value = response.data;
			} catch (error) {
				console.error("Failed to fetch user:", error);
			}
		};

		const fetchUsers = async () => {
			const userId = localStorage.getItem("userId");

			try {
				const response = await axios.get("/users");
				allUsers.value = response.data.filter(
					(u) => u.id !== Number(userId)
				);
			} catch (error) {
				console.error("Failed to fetch users:", error);
			}
		};

		const fetchConversations = async () => {
			try {
				const response = await axios.get("/conversations");
				conversations.value = response.data.conversations;
			} catch (error) {
				console.error("Failed to fetch conversations:", error);
			}
		};

		const selectConversation = (conversation) => {
			selectedConversation.value = conversation;
		};

		const logout = () => {
			localStorage.removeItem("userId");
			router.push("/login");
		};

		const handleFeedback = (message) => {
			feedbackMessage.value = message;
			showFeedback.value = true;

			setTimeout(() => {
				showFeedback.value = false;
			}, 3000);
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

		const updatePhoto = async (file) => {
			const formData = new FormData();
			formData.append("photo", file);

			try {
				const response = await axios.put("/user/photo", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				console.log("Profile picture updated:", response.data);
				user.value.photo_url = response.data.photo_url;
				return { success: true };
			} catch (error) {
				console.error(
					"Failed to update profile picture:",
					error.response?.data || error.message
				);
				return {
					success: false,
					error: "Failed to upload profile picture.",
				};
			}
		};

		const updateConversationName = (payload) => {
			const conversation = conversations.value.find(
				(conv) => conv.conversation_id === payload.conversationId
			);
			if (conversation) {
				conversation.display_name = payload.newName;
			}
		};

		const updateConversationPhoto = (payload) => {
			if (
				selectedConversation.value &&
				selectedConversation.value.conversation_id ===
					payload.conversationId
			) {
				selectedConversation.value.display_photo_url =
					payload.newPhotoUrl;
			}

			const conversationIndex = conversations.value.findIndex(
				(conversation) =>
					conversation.conversation_id === payload.conversationId
			);

			if (conversationIndex !== -1) {
				conversations.value[conversationIndex].display_photo_url =
					payload.newPhotoUrl;
			}
		};

		onMounted(async () => {
			await fetchUser();
			await fetchUsers();
			await fetchConversations();

			if (conversations.value.length > 0) {
				selectedConversation.value = conversations.value[0];
				console.log(
					"Selected conversation:",
					selectedConversation.value
				);
			}
		});

		return {
			logout,
			updateUsername,
			updatePhoto,
			user,
			allUsers,
			conversations,
			handleFeedback,
			feedbackMessage,
			showFeedback,
			selectConversation,
			selectedConversation,
			updateConversationPhoto,
			updateConversationName,
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
					:updatePhoto="updatePhoto"
				/>
			</div>
			<div class="col-auto">
				<ConversationList
					v-if="user.id"
					@feedback="handleFeedback"
					@select-conversation="selectConversation"
					:user="user"
					:conversations="conversations"
					:allUsers="allUsers"
					:selectedConversation="selectedConversation"
				/>
			</div>
			<div class="col">
				<Conversation
					v-if="selectedConversation"
					:conversation="selectedConversation"
					:user="user"
					@group-photo-updated="updateConversationPhoto"
					@group-name-updated="updateConversationName"
				/>
			</div>
		</div>
	</div>
</template>
