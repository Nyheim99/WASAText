<script>
import Avatar from "./Avatar.vue";

export default {
	props: {
		logout: Function,
		updateUsername: Function,
		updatePhoto: Function,
		user: {
			type: Object,
			required: true,
		},
	},
	components: {
		Avatar,
	},
};
</script>

<template>
	<div class="h-100 rounded d-flex flex-column justify-content-end">
		<Avatar
			:logout="logout"
			:user="user"
			:updateUsername="updateUsername"
			:updatePhoto="updatePhoto"
		/>
	</div>
</template>
