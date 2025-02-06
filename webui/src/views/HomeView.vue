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
		const selectedConversationDetails = ref(null);
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
			try {
				const response = await axios.get("/users");
				allUsers.value = response.data.filter((u) => u.id !== user.id);
			} catch (error) {
				console.error("Failed to fetch users:", error);
			}
		};

		const fetchConversations = async () => {
			try {
				const response = await axios.get("/conversations");
				conversations.value = response.data;
				console.log(
					"Fetched conversations in HomeView.vue: ",
					response.data
				);
			} catch (error) {
				console.error("Failed to fetch conversations:", error);
			}
		};

		const selectConversation = (conversation) => {
			selectedConversation.value = conversation;
			fetchConversationDetails(conversation.conversation_id);
		};

		const fetchConversationDetails = async (conversationId) => {
			try {
				const response = await axios.get(
					`/conversations/${conversationId}`
				);
				selectedConversationDetails.value = response.data;
				console.log("Conversation details in HomeView:", response.data);
			} catch (error) {
				console.error("Failed to fetch conversation details:", error);
			}
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
				await axios.put("/user/username", {
					username: newUsername,
				});
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
				return true;
			} catch (error) {
				console.error(
					"Failed to update profile picture:",
					error.response?.data || error.message
				);
				return {
					error: "Failed to upload profile picture.",
				};
			}
		};

		const addNewConversation = async () => {
			try {
				await fetchConversations();

				if (conversations.value.length > 0) {
					selectConversation(conversations.value[0]);
				}
				
			} catch (error) {
				console.error("Failed to refetch conversations:", error);
			}
		};

		const removeConversation = (conversationId) => {
			conversations.value = conversations.value.filter(
				(conv) => conv.conversation_id !== conversationId
			);
			selectedConversation.value = null;
			selectedConversationDetails.value = null;

			if (conversations.value.length > 0) {
				selectConversation(conversations.value[0]);
			}
		};

		const updateConversationName = (payload) => {
			const index = conversations.value.findIndex(
				(conv) => conv.conversation_id === payload.conversationId
			);

			if (index !== -1) {
				conversations.value[index] = {
					...conversations.value[index],
					display_name: payload.newName,
				};

				if (
					selectedConversation.value?.conversation_id ===
					payload.conversationId
				) {
					selectedConversation.value = {
						...selectedConversation.value,
						display_name: payload.newName,
					};
				}
			}
		};

		const updateConversationPhoto = (payload) => {
			const index = conversations.value.findIndex(
				(conv) => conv.conversation_id === payload.conversationId
			);

			if (index !== -1) {
				conversations.value[index] = {
					...conversations.value[index],
					display_photo_url: payload.newPhotoUrl,
				};
			}

			if (
				selectedConversation.value?.conversation_id ===
				payload.conversationId
			) {
				selectedConversation.value = {
					...selectedConversation.value,
					display_photo_url: payload.newPhotoUrl,
				};
			}
		};

		const updateConversationWithNewMessage = (payload) => {
			const { conversationId, lastMessage } = payload;

			console.log(lastMessage);

			fetchConversations();
			fetchConversationDetails(conversationId);
		};

		const updateConversationWithDeletedMessage = (payload) => {
			const { conversationId, deletedMessage } = payload;

			console.log(deletedMessage);

			fetchConversations();
			fetchConversationDetails(conversationId);
		};

		onMounted(async () => {
			await fetchUser();
			await fetchUsers();
			await fetchConversations();

			if (conversations.value.length > 0) {
				selectConversation(conversations.value[0]);
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
			selectedConversationDetails,
			updateConversationPhoto,
			updateConversationName,
			updateConversationWithNewMessage,
			updateConversationWithDeletedMessage,
			fetchConversationDetails,
			addNewConversation,
			removeConversation,
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

		<div class="row flex-grow-1 g-3 flex-nowrap h-100">
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
					@conversation-created="addNewConversation"
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
					:conversationDetails="selectedConversationDetails"
					:user="user"
					:allUsers="allUsers"
					@group-photo-updated="updateConversationPhoto"
					@group-name-updated="updateConversationName"
					@group-members-updated="fetchConversationDetails"
					@group-left="removeConversation"
					@message-sent="updateConversationWithNewMessage"
					@message-deleted="updateConversationWithDeletedMessage"
				/>
			</div>
		</div>
	</div>
</template>
