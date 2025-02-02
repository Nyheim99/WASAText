<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
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
		allUsers: {
			type: Array,
			required: true,
		},
	},
	emits: [
		"group-photo-updated",
		"group-name-updated",
		"group-members-updated",
		"group-left",
	],
	setup(props, { emit }) {
		const fileInput = ref(null);
		const uploading = ref(false);
		const validationMessage = ref("");
		const cacheBuster = ref(Date.now());
		const newGroupName = ref("");
		const updatingName = ref(false);

		const usersNotInGroup = ref([]);
		const selectedUsers = ref(new Set());
		const addingMembers = ref(false);

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

		onBeforeUnmount(() => {
			const modal = document.getElementById("groupNameModal");
			if (modal) {
				modal.removeEventListener(
					"shown.bs.modal",
					setGroupNameOnModalOpen
				);
			}
		});

		onMounted(() => {
			updateUsersNotInGroup();
			const modal = document.getElementById("groupNameModal");
			if (modal) {
				modal.addEventListener(
					"shown.bs.modal",
					setGroupNameOnModalOpen
				);
			}
		});

		const updateUsersNotInGroup = () => {
			if (
				!props.conversationDetails ||
				!props.conversationDetails.participants
			) {
				usersNotInGroup.value = [];
				return;
			}

			const groupMemberIds = new Set(
				props.conversationDetails.participants.map((u) => u.id)
			);

			usersNotInGroup.value = props.allUsers.filter(
				(user) => !groupMemberIds.has(user.id)
			);
		};

		const toggleUserSelection = (userId) => {
			if (selectedUsers.value.has(userId)) {
				selectedUsers.value.delete(userId);
			} else {
				selectedUsers.value.add(userId);
			}
		};

		const getUsername = (userId) => {
			const user = props.allUsers.find((user) => user.id === userId);
			return user ? user.username : "Unknown";
		};

		const handleAddMembers = async () => {
			if (selectedUsers.value.size === 0) {
				alert("Please select at least one user.");
				return;
			}

			addingMembers.value = true;

			try {
				// Send API request to add members
				const response = await axios.post(
					`/conversations/${props.conversation.conversation_id}/members`,
					{
						participants: [...selectedUsers.value],
					}
				);

				console.log("Added members:", response.data);

				emit(
					"group-members-updated",
					props.conversation.conversation_id
				);

				// Refresh the conversation details
				await fetchConversationDetails(
					props.conversation.conversation_id
				);

				// Reset selection
				selectedUsers.value.clear();
				updateUsersNotInGroup();

				// Close the modal
				const modal = document.getElementById("addMembersModal");
				if (modal) {
					const bootstrapModal = bootstrap.Modal.getInstance(modal);
					if (bootstrapModal) {
						bootstrapModal.hide();
					}
				}
			} catch (error) {
				console.error("Failed to add members:", error);
				alert(
					error.response?.data?.message || "Failed to add members."
				);
			} finally {
				addingMembers.value = false;
			}
		};

		const fetchConversationDetails = async (conversationId) => {
			try {
				const response = await axios.get(
					`/conversations/${conversationId}`
				);
				props.conversationDetails.participants =
					response.data.participants;
				console.log("Updated conversation details:", response.data);
			} catch (error) {
				console.error(
					"Failed to fetch updated conversation details:",
					error
				);
			}
		};

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

		const handleLeaveGroup = async () => {
			try {
				await axios.delete(
					`/conversations/${props.conversation.conversation_id}/leave`
				);

				emit("group-left", props.conversation.conversation_id);

				// Close the modal manually
				const modalElement = document.getElementById("leaveGroupModal");
				const modal = bootstrap.Modal.getInstance(modalElement);
				if (modal) modal.hide();

				// Clear conversation details (remove from UI)
				props.conversationDetails = null;
				console.log("Left group successfully");
			} catch (error) {
				console.error("Failed to leave group:", error);
				alert(
					error.response?.data?.message || "Failed to leave group."
				);
			}
		};

		const resolvePhotoURL = (photoURL) => {
			if (!photoURL) {
				return AvatarIcon;
			}
			if (photoURL.startsWith("/")) {
				return `${__API_URL__}${photoURL}`;
			}
			return photoURL;
		};

		watch(
			() => props.conversationDetails,
			() => {
				updateUsersNotInGroup();
				selectedUsers.value.clear();
			},
			{ deep: true }
		);

		return {
			fileInput,
			uploading,
			newGroupName,
			updatingName,
			validationMessage,
			conversationPhoto,
			handleUpdateGroupPhoto,
			handleUpdateGroupName,
			handleAddMembers,
			handleLeaveGroup,
			getUsername,
			toggleUserSelection,
			selectedUsers,
			addingMembers,
			usersNotInGroup,
			resolvePhotoURL,
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
			<button
				v-if="conversation.conversation_type === 'group'"
				class="btn btn-outline-primary"
				data-bs-toggle="modal"
				data-bs-target="#addMembersModal"
			>
				Add Members
			</button>
			<button
				v-if="conversation.conversation_type === 'group'"
				class="btn btn-outline-danger"
				data-bs-toggle="modal"
				data-bs-target="#leaveGroupModal"
			>
				Leave Group
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
	<div
		class="modal fade"
		id="addMembersModal"
		tabindex="-1"
		aria-labelledby="addMembersModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h1 class="modal-title fs-5" id="addMembersModalLabel">
						Add Members to Group
					</h1>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
						aria-label="Close"
					></button>
				</div>
				<div class="modal-body">
					<div class="dropdown">
						<button
							class="btn btn-light dropdown-toggle"
							type="button"
							id="selectUsersDropdown"
							data-bs-toggle="dropdown"
							aria-expanded="false"
						>
							Select Users
						</button>
						<ul
							class="dropdown-menu"
							aria-labelledby="selectUsersDropdown"
							style="max-height: 300px; overflow-y: auto"
						>
							<li
								v-for="user in usersNotInGroup"
								:key="user.id"
								class="dropdown-item d-flex align-items-center justify-content-between"
							>
								<div class="d-flex align-items-center">
									<img
										:src="resolvePhotoURL(user.photo_url)"
										alt="User Avatar"
										class="rounded-circle me-2"
										style="
											width: 30px;
											height: 30px;
											object-fit: cover;
										"
									/>
									{{ user.username }}
								</div>
								<input
									type="checkbox"
									:checked="selectedUsers.has(user.id)"
									@click="toggleUserSelection(user.id)"
								/>
							</li>
						</ul>
					</div>

					<!-- Selected users display -->
					<div class="mt-3">
						<h6>Selected Users:</h6>
						<div class="d-flex flex-wrap gap-2">
							<span
								v-for="userId in selectedUsers"
								:key="userId"
								class="badge text-bg-secondary d-flex align-items-center"
								style="font-size: 14px; padding: 0.5em 0.75em"
							>
								{{ getUsername(userId) }}
								<button
									class="btn-close btn-close-white ms-2"
									aria-label="Remove"
									@click="toggleUserSelection(userId)"
									style="font-size: 10px; opacity: 0.8"
								></button>
							</span>
						</div>
					</div>
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
						@click="handleAddMembers"
						:disabled="addingMembers"
					>
						{{ addingMembers ? "Adding..." : "Add to Group" }}
					</button>
				</div>
			</div>
		</div>
	</div>
	<!-- Leave Group Confirmation Modal -->
	<div
		class="modal fade"
		id="leaveGroupModal"
		tabindex="-1"
		aria-labelledby="leaveGroupModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="leaveGroupModalLabel">
						Leave Group
					</h5>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
						aria-label="Close"
					></button>
				</div>
				<div class="modal-body">
					<p>Are you sure you want to leave this group?</p>
				</div>
				<div class="modal-footer">
					<button
						type="button"
						class="btn btn-secondary"
						data-bs-dismiss="modal"
					>
						Cancel
					</button>
					<button
						type="button"
						class="btn btn-danger"
						@click="handleLeaveGroup"
					>
						Leave Group
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
