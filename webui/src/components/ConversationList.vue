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
	<div class="conversation-list-panel">
		<h2>Conversations</h2>
		<ul v-if="conversations.length > 0">
			<li
				v-for="conversation in conversations"
				:key="conversation.conversationId"
			>
				{{ conversation.name }}
			</li>
		</ul>
		<p v-else class="no-conversations">
			No conversations found. Start a new chat!
		</p>
	</div>
</template>

<style>
.conversation-list-panel {
	background-color: #ecf0f1;
	padding: 20px;
	overflow-y: auto;
}

.conversation-list {
	list-style: none;
	padding: 0;
	margin: 0;
}

.conversation-item {
	padding: 10px;
	margin-bottom: 10px;
	background-color: #bdc3c7;
	border-radius: 5px;
	cursor: pointer;
}

.conversation-item:hover {
	background-color: #95a5a6;
}

.no-conversations {
  font-style: italic;
  color: #888;
  text-align: center;
}
</style>
