<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";
import { ref, onMounted, onBeforeUnmount } from "vue";
import axios from "../services/axios";

export default {
	props: {
		conversation: {
			type: Object,
			required: true,
		},
		conversationDetails: {
			type: Object,
			required: false,
			default: null,
		},
		user: {
			type: Object,
			required: true,
		},
	},
	emits: ["group-photo-updated", "group-name-updated"],
	setup(props, { emit }) {
		const fileInput = ref(null);
		const uploading = ref(false);
		const validationMessage = ref("");
		const cacheBuster = ref(Date.now());
		const newGroupName = ref("");
		const updatingName = ref(false);

		const conversationPhoto = () => {
			if (props.conversation.display_photo_url?.startsWith("/")) {
				return `${__API_URL__}${props.conversation.display_photo_url}?t=${cacheBuster.value}`;
			}
			return props.conversation.conversation_type === "group"
				? PeopleIcon
				: AvatarIcon;
		};

		const validateImage = (file) => {
			const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
			if (!allowedTypes.includes(file.type)) {
				validationMessage.value =
					"Only JPEG and PNG files are allowed.";
				return false;
			}
			return true;
		};

		const setGroupNameOnModalOpen = () => {
			newGroupName.value = props.conversation.display_name || "";
		};

		onMounted(() => {
			const modal = document.getElementById("groupNameModal");
			if (modal) {
				modal.addEventListener(
					"shown.bs.modal",
					setGroupNameOnModalOpen
				);
			}
		});

		onBeforeUnmount(() => {
			const modal = document.getElementById("groupNameModal");
			if (modal) {
				modal.removeEventListener(
					"shown.bs.modal",
					setGroupNameOnModalOpen
				);
			}
		});

		const handleUpdateGroupName = async () => {
			if (!newGroupName.value.trim()) {
				validationMessage.value = "Group name cannot be empty.";
				return;
			}

			if (
				newGroupName.value.length < 3 ||
				newGroupName.value.length > 20
			) {
				validationMessage.value =
					"Group name must be between 3 and 20 characters.";
				return;
			}

			updatingName.value = true;

			try {
				const response = await axios.put(
					`/conversations/${props.conversation.conversation_id}/name`,
					{
						name: newGroupName.value,
					}
				);

				console.log("Group name updated:", response.data);

				props.conversation.display_name = response.data.name;

				emit("group-name-updated", {
					conversationId: props.conversation.conversation_id,
					newName: response.data.name,
				});

				const modal = document.getElementById("groupNameModal");
				if (modal) {
					const bootstrapModal = bootstrap.Modal.getInstance(modal);
					if (bootstrapModal) {
						bootstrapModal.hide();
					}
				}

				newGroupName.value = "";
				validationMessage.value = "";
			} catch (error) {
				console.error("Failed to update group name:", error);
				validationMessage.value =
					error.response?.data?.message ||
					"Failed to update group name.";
			} finally {
				updatingName.value = false;
			}
		};

		const handleUpdateGroupPhoto = async () => {
			const file = fileInput.value?.files[0];
			if (!file) {
				validationMessage.value = "Please select a file.";
				return;
			}

			if (!validateImage(file)) {
				return;
			}

			uploading.value = true;
			const formData = new FormData();
			formData.append("photo", file);

			try {
				const response = await axios.put(
					`/conversations/${props.conversation.conversation_id}/photo`,
					formData,
					{
						headers: { "Content-Type": "multipart/form-data" },
					}
				);

				console.log("Group picture updated:", response.data);

				cacheBuster.value = Date.now();
				props.conversation.display_photo_url =
					response.data.photo_url + "?t=" + cacheBuster.value;

				emit("group-photo-updated", {
					conversationId: props.conversation.conversation_id,
					newPhotoUrl: props.conversation.display_photo_url,
				});

				if (fileInput.value) {
					fileInput.value.value = "";
				}

				const modal = document.getElementById("groupPhotoModal");
				if (modal) {
					const bootstrapModal = bootstrap.Modal.getInstance(modal);
					if (bootstrapModal) {
						bootstrapModal.hide();
					}
				}

				uploading.value = false;
			} catch (error) {
				console.error("Failed to update group photo:", error);
				validationMessage.value =
					error.response?.data?.message ||
					"Failed to upload group picture.";
				uploading.value = false;
			}
		};

		return {
			fileInput,
			uploading,
			newGroupName,
			updatingName,
			validationMessage,
			conversationPhoto,
			handleUpdateGroupPhoto,
			handleUpdateGroupName,
		};
	},
};
</script>

<template>
	<div class="bg-white shadow-sm rounded p-4 overflow-auto h-100">
		<div class="d-flex align-items-center justify-content-between mb-2">
			<div class="d-flex align-items-center">
				<img
					:src="conversationPhoto()"
					alt="Conversation Avatar"
					class="rounded-circle"
					style="width: 50px; height: 50px; object-fit: cover"
				/>
				<h2 class="px-2 mb-0">{{ conversation.display_name }}</h2>
			</div>

			<button
				v-if="conversation.conversation_type === 'group'"
				class="btn btn-outline-primary ms-2"
				data-bs-toggle="modal"
				data-bs-target="#groupNameModal"
			>
				Update Group Name
			</button>

			<button
				v-if="conversation.conversation_type === 'group'"
				class="btn btn-outline-primary"
				data-bs-toggle="modal"
				data-bs-target="#groupPhotoModal"
			>
				Update Group Picture
			</button>
		</div>

		<div
			class="modal fade"
			id="groupNameModal"
			tabindex="-1"
			aria-labelledby="groupNameModalLabel"
			aria-hidden="true"
		>
			<div class="modal-dialog">
				<div class="modal-content">
					<div class="modal-header">
						<h1 class="modal-title fs-5" id="groupNameModalLabel">
							Update Group Name
						</h1>
						<button
							type="button"
							class="btn-close"
							data-bs-dismiss="modal"
							aria-label="Close"
						></button>
					</div>
					<div class="modal-body">
						<input
							type="text"
							v-model="newGroupName"
							class="form-control"
							placeholder="Enter new group name"
						/>
						<p
							class="text-danger small mt-2"
							v-if="validationMessage"
						>
							{{ validationMessage }}
						</p>
					</div>
					<div class="modal-footer">
						<button
							type="button"
							class="btn btn-secondary"
							data-bs-dismiss="modal"
						>
							Close
						</button>
						<button
							type="button"
							class="btn btn-primary"
							@click="handleUpdateGroupName"
							:disabled="updatingName"
						>
							{{ updatingName ? "Updating..." : "Update" }}
						</button>
					</div>
				</div>
			</div>
		</div>

		<hr class="mx-n4" />
		<p>Conversation ID: {{ conversation.conversation_id }}</p>
	</div>

	<div
		class="modal fade"
		id="groupPhotoModal"
		tabindex="-1"
		aria-labelledby="groupPhotoModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h1 class="modal-title fs-5" id="groupPhotoModalLabel">
						Upload a new group picture
					</h1>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
						aria-label="Close"
					></button>
				</div>
				<div class="modal-body">
					<div class="input-group flex-nowrap">
						<input
							type="file"
							ref="fileInput"
							accept="image/jpeg, image/png"
							class="form-control"
						/>
					</div>
					<p class="text-danger small mt-2" v-if="validationMessage">
						{{ validationMessage }}
					</p>
				</div>
				<div class="modal-footer">
					<button
						type="button"
						class="btn btn-secondary"
						data-bs-dismiss="modal"
					>
						Close
					</button>
					<button
						type="button"
						class="btn btn-primary"
						@click="handleUpdateGroupPhoto"
						:disabled="uploading"
					>
						{{ uploading ? "Uploading..." : "Upload" }}
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
