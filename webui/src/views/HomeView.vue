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

		//Fetch user details
		const fetchUser = async () => {
			try {
				const response = await axios.get("/user");

				//Set the user object on success
				user.value = response.data;
			} catch (error) {
				console.error("Failed to fetch user:", error);
			}
		};

		//Fetch all users
		const fetchUsers = async () => {
			try {
				const response = await axios.get("/users");
				allUsers.value = response.data.filter(
					(u) => u.id !== user.value.id
				);
			} catch (error) {
				console.error("Failed to fetch users:", error);
			}
		};

		//Fetch all conversations
		const fetchConversations = async () => {
			try {
				const response = await axios.get("/conversations");
				conversations.value = response.data;
			} catch (error) {
				console.error("Failed to fetch conversations:", error);
			}
		};

		//Select a conversation
		const selectConversation = async (conversation) => {
			selectedConversation.value = conversation;
			await markMessagesAsRead(conversation.conversation_id);
			fetchConversationDetails(conversation.conversation_id);
		};

		//Fetch details about a single conversation
		const fetchConversationDetails = async (conversationId) => {
			try {
				const response = await axios.get(
					`/conversations/${conversationId}`
				);
				selectedConversationDetails.value = response.data;
			} catch (error) {
				console.error("Failed to fetch conversation details:", error);
			}
		};

		//Logging out
		const logout = () => {
			localStorage.removeItem("userId");
			router.push("/login");
		};

		//Handling feedback messages from other components
		const handleFeedback = (message) => {
			feedbackMessage.value = message;
			showFeedback.value = true;

			setTimeout(() => {
				showFeedback.value = false;
			}, 3000);
		};

		//Updating a user's username
		const updateUsername = async (newUsername) => {
			try {
				//Attempt to update username
				await axios.put("/user/username", {
					username: newUsername,
				});

				//Set feedback on success
				feedbackMessage.value = "Username updated successfully!";
				showFeedback.value = true;

				//Optimistically update the username locally
				user.value.username = newUsername;

				//setTimeout for feedback
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

		//Update a user's profile picture
		const updatePhoto = async (file) => {
			const formData = new FormData();
			formData.append("photo", file);

			try {
				//Attempt to update photo
				const response = await axios.put("/user/photo", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				//Optimistically update the picture locally
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

		//Updating conversations after adding a new one
		const addNewConversation = async () => {
			try {
				//Fetching details about all conversations
				await fetchConversations();

				//As long as one exist, select the newest one
				if (conversations.value.length > 0) {
					selectConversation(conversations.value[0]);
				}
			} catch (error) {
				console.error("Failed to refetch conversations:", error);
			}
		};

		//Remove a conversation after the user leaves a group
		const removeConversation = (conversationId) => {
			//Remove the conversation locally
			conversations.value = conversations.value.filter(
				(conv) => conv.conversation_id !== conversationId
			);

			//Remove the selected conversation
			selectedConversation.value = null;
			selectedConversationDetails.value = null;

			//If the user still has conversations remaining, select it
			if (conversations.value.length > 0) {
				selectConversation(conversations.value[0]);
			}
		};

		//Change the name of the group conversation after updating it
		const updateConversationName = (payload) => {
			//Find the index of the conversation
			const index = conversations.value.findIndex(
				(conv) => conv.conversation_id === payload.conversationId
			);

			//If the index is valid, update display name
			if (index !== -1) {
				conversations.value[index] = {
					...conversations.value[index],
					display_name: payload.newName,
				};

				//If the user is in the selected conversation, also updates its name
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

		//Update the group photo when the user uploads a new one
		const updateConversationPhoto = (payload) => {
			//Find the index of the conversation
			const index = conversations.value.findIndex(
				(conv) => conv.conversation_id === payload.conversationId
			);

			//if the index is valid, update the display photo url
			if (index !== -1) {
				conversations.value[index] = {
					...conversations.value[index],
					display_photo_url: payload.newPhotoUrl,
				};
			}

			//Also update for the selected conversation
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

		//Update conversation preview and conversation with new message
		const updateConversationWithNewMessage = (conversationId) => {
			//Get all conversations to update the preview
			fetchConversations();

			//Get the details of the current conversation to update with new message
			fetchConversationDetails(conversationId);
		};

		//Update conversation preview and conversation with new message
		const updateConversationWithDeletedMessage = (conversationId) => {
			//Get all conversations to update the preview
			fetchConversations();

			//Get all the details of the current conversation to update with deleted message
			fetchConversationDetails(conversationId);
		};

		//Mark all messages as read for a user in a conversation
		const markMessagesAsRead = async (conversationId) => {
			try {
				await axios.put(
					`/conversations/${conversationId}/messages/read`
				);
			} catch (error) {
				console.error("Failed to mark messages as read:", error);
			}
		};

		//Fetch all data in component mount
		onMounted(async () => {
			await fetchUser();
			await fetchUsers();
			await fetchConversations();

			//Automatically select the newest conversation if one exists
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
		<!-- Feedback box-->
		<div
			v-if="showFeedback"
			class="alert alert-success position-fixed top-2 start-50 translate-middle-x shadow"
			role="alert"
			style="z-index: 1050"
		>
			{{ feedbackMessage }}
		</div>

		<!-- Main window -->
		<div class="row flex-grow-1 g-3 flex-nowrap h-100">

			<!-- Sidebar -->
			<div class="col-auto p-0">
				<Sidebar
					:logout="logout"
					:user="user"
					:updateUsername="updateUsername"
					:updatePhoto="updatePhoto"
				/>
			</div>

			<!-- Conversation List -->
			<div class="col-auto">
				<ConversationList
					v-if="user.id"
					:user="user"
					:conversations="conversations"
					:allUsers="allUsers"
					:selectedConversation="selectedConversation"
					@feedback="handleFeedback"
					@select-conversation="selectConversation"
					@conversation-created="addNewConversation"
				/>
			</div>

			<!-- Conversation Window -->
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
					@message-forwarded="updateConversationWithNewMessage"
					@message-deleted="updateConversationWithDeletedMessage"
					@conversation-created="addNewConversation"
				/>

				<!-- If user has no conversations, display an empty box -->
				<div
					v-else
					class="bg-white d-flex flex-column shadow-sm rounded h-100"
				></div>
			</div>
		</div>
	</div>
</template>
