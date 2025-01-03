<script>
export default {
	props: {
		conversations: {
			type: Array,
			required: true,
            default: () => [],
		},
	},
};
</script>

<template>
  <div class="rounded shadow-sm bg-white p-3 overflow-auto h-100">
    <h2 class="mb-3">Conversations</h2>
    <ul v-if="conversations.length > 0" class="list-unstyled">
      <li
        v-for="conversation in conversations"
        :key="conversation.conversationId"
        class="p-2 mb-2 bg-secondary text-white rounded"
      >
        {{ conversation.name }}
      </li>
    </ul>
    <p v-else class="text-muted text-center">
      No conversations found. Start a new chat!
    </p>
  </div>
</template>