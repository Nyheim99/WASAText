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
  <div class="home-container">
    <Sidebar :logout="logout" />

	<ConversationList :conversations="conversations" />

    <div class="chat-main-panel">
      <h2>Chat Window</h2>
      <p>Select a chat to start messaging!</p>
    </div>
  </div>
</template>

<style>
.home-container {
  display: grid;
  grid-template-columns: 1fr 2fr 6fr;
  height: calc(100vh - 40px);
  box-sizing: border-box;
}

.chat-main-panel {
  background-color: #ffffff;
  padding: 20px;
  overflow-y: auto;
}
</style>
