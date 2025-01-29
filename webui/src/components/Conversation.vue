<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";

export default {
	props: {
		conversation: {
			type: Object,
			required: true,
		},
	},

	mounted() {
		console.log("Received conversation:", this.conversation);
	},
	computed: {
		conversationPhoto() {
			if (this.conversation.display_photo_url?.startsWith("/")) {
				return `${__API_URL__}${this.conversation.display_photo_url}`;
			}
			return this.conversation.conversation_type === "group"
				? PeopleIcon
				: AvatarIcon;
		},
	},
};
</script>

<template>
	<div class="bg-white shadow-sm rounded p-4 overflow-auto h-100">
		<div class="d-flex align-items-center mb-2">
			<img
				:src="conversationPhoto"
				alt="Conversation Avatar"
				class="rounded-circle"
				style="width: 50px; height: 50px; object-fit: cover"
			/>
			<h2 class="px-2">{{ conversation.display_name }}</h2>
		</div>
		<hr class="mx-n4" />
		<p>Conversation ID: {{ conversation.conversation_id }}</p>
	</div>
</template>
