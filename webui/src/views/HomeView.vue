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

		const logout = () => {
			localStorage.removeItem("userId");
			router.push("/login");
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
			fetchConversations();
		});

		return {
			logout,
			conversations,
		};
	},
};
</script>

<template>
	<div class="container-fluid d-flex vh-100 flex-column p-3">
		<div class="row flex-grow-1 g-3">
			<div class="col-auto">
				<Sidebar :logout="logout" />
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
