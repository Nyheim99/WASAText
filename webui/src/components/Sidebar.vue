<script>
import Avatar from "./Avatar.vue";
import AvatarIcon from "../assets/person-circle.svg";

export default {
  props: {
    logout: Function,
	updateUsername: Function,
	user: {
		type: Object,
		required: true
	}
  },
  components: {
    Avatar,
  },
  setup() {
    return {
      AvatarIcon,
    };
  },
};
</script>

<template>
	<div class="h-100 rounded d-flex flex-column justify-content-end">
		<Avatar :src="AvatarIcon" :logout="logout" :updateUsername="updateUsername" :user="user"/>
	</div>
</template>
