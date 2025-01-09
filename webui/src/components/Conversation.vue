<script>
export default {
	props: {
		conversationId: {
			type: Number,
			required: true,
		},
	},
};
</script>

<template>
	<div class="bg-white shadow-sm rounded p-4 overflow-auto h-100">
		<h2>Conversation</h2>
		<p>Conversation ID: {{ conversationId }}</p>
	</div>
</template>
