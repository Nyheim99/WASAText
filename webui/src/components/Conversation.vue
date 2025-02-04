<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";
import ImageIcon from "/image.svg";
import SendIcon from "/send.svg";
import { ref, onMounted, onBeforeUnmount, watch, computed, nextTick } from "vue";
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
		"message-sent",
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

		const newMessage = ref("");
		const photoInput = ref(null);
		const photoPreviewDiv = ref(null);
		const selectedPhoto = ref(null);
		const messages = computed(
			() => props.conversationDetails?.messages || []
		);

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

		const sendMessage = async () => {
			if (!newMessage.value && !selectedPhoto.value) {
				alert("Please enter a message or select a photo.");
				return;
			}

			const formData = new FormData();
			formData.append("message", newMessage.value);
			if (selectedPhoto.value) {
				formData.append("photo", selectedPhoto.value);
			}

			try {
				const response = await axios.post(
					`/conversations/${props.conversation.conversation_id}/messages`,
					formData,
					{ headers: { "Content-Type": "multipart/form-data" } }
				);

				console.log("Message sent:", response.data);

				newMessage.value = "";
				selectedPhoto.value = null;

				emit("message-sent", {
					conversationId: props.conversation.conversation_id,
					lastMessage: response.data,
				});
			} catch (error) {
				console.error(
					"Failed to send message:",
					error.response?.data || error.message
				);
			}
		};

		const triggerFileUpload = () => {
			photoInput.value.click();
		};

		const photoPreview = computed(() => {
			return selectedPhoto.value
				? URL.createObjectURL(selectedPhoto.value)
				: null;
		});

		const handlePhotoUpload = (event) => {
			const file = event.target.files[0];
			if (file) {
				selectedPhoto.value = file;

				nextTick(() => {
            photoPreviewDiv.value?.focus();
        });
			}
		};

		const formatBase64Image = (base64Data, mimeType) => {
			if (!base64Data) return "";
			return `data:${mimeType};base64,${base64Data}`;
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

		const getSenderPhoto = (senderId) => {
			const sender = props.allUsers.find((user) => user.id === senderId);
			return sender?.photo_url || "";
		};

		const formatTimestamp = (timestamp) => {
			const date = new Date(timestamp);
			return date.toLocaleTimeString("en-GB", {
				hour: "2-digit",
				minute: "2-digit",
				hour12: false,
			});
		};

		watch(
			() => props.conversationDetails,
			() => {
				updateUsersNotInGroup();
				selectedUsers.value.clear();
			},
			{ deep: true }
		);

		const handleEnterPress = (event) => {
			event.preventDefault();
			console.log("Enter pressed")

			if (selectedPhoto.value || newMessage.value.trim()) {
				sendMessage();
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
			handleAddMembers,
			handleLeaveGroup,
			getUsername,
			toggleUserSelection,
			selectedUsers,
			addingMembers,
			usersNotInGroup,
			resolvePhotoURL,
			formatTimestamp,
			messages,
			getSenderPhoto,
			sendMessage,
			handlePhotoUpload,
			triggerFileUpload,
			newMessage,
			photoInput,
			SendIcon,
			ImageIcon,
			formatBase64Image,
			selectedPhoto,
			photoPreview,
			photoPreviewDiv,
			handleEnterPress
		};
	},
};
</script>

<template>
	<div class="bg-white d-flex flex-column shadow-sm rounded p-4 h-100">
		<div class="d-flex align-items-center mb-2">
			<div class="d-flex align-items-center">
				<img
					:src="conversationPhoto()"
					alt="Conversation Avatar"
					class="rounded-circle"
					style="width: 50px; height: 50px; object-fit: cover"
				/>
				<h2 class="px-2 mb-0">
					{{ conversation.display_name }} ({{
						conversation.conversation_id
					}})
				</h2>
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

		<hr class="m-0" />

		<div class="flex-grow-1 overflow-auto p-3 d-flex flex-column">
			<div
				v-for="message in messages"
				:key="message.id"
				class="d-flex align-items-start mb-2"
				:style="{
					justifyContent:
						message.sender_id === user.id
							? 'flex-end'
							: 'flex-start',
				}"
				style="position: relative"
			>
				<!-- Avatar for received messages -->
				<img
					v-if="message.sender_id !== user.id"
					:src="resolvePhotoURL(getSenderPhoto(message.sender_id))"
					alt="Sender Avatar"
					class="rounded-circle me-2"
					style="
						width: 35px;
						height: 35px;
						object-fit: cover;
						align-self: flex-end;
					"
				/>

				<!-- Message Bubble -->
				<div
					class="p-2 rounded shadow-sm"
					:style="{
						backgroundColor:
							message.sender_id === user.id
								? '#dcf8c6'
								: '#f1f0f0',
						maxWidth: '75%',
						wordWrap: 'break-word',
						padding: '10px',
						borderRadius: '18px',
						position: 'relative',
						alignSelf:
							message.sender_id === user.id
								? 'flex-end'
								: 'flex-start',
						textAlign: 'left',
					}"
				>
					<p
						v-if="
							conversation.conversation_type === 'group' &&
							message.sender_id !== user.id
						"
						style="
							font-weight: bold;
							margin-bottom: 2px;
							font-size: 14px;
						"
					>
						{{ message.sender_username }}
					</p>
					<div
						style="
							display: flex;
							flex-direction: column;
							position: relative;
						"
					>
						<p
							v-if="message.content"
							class="mb-0"
							style="
								margin: 0;
								word-break: break-word;
								white-space: pre-wrap;
							"
						>
							{{ message.content }}
						</p>
						<!-- Image Message -->
						<img
							v-if="message.photo_data"
							:src="
								formatBase64Image(
									message.photo_data,
									message.photo_mime_type
								)
							"
							alt="Sent Image"
							class="mt-2 rounded"
							style="max-width: 200px; border-radius: 8px"
						/>
						<small
							class="text-muted"
							style="
								font-size: 12px;
								color: gray;
								align-self: flex-end;
								margin-top: 4px;
								white-space: nowrap;
							"
						>
							{{ formatTimestamp(message.timestamp) }}
						</small>
					</div>
				</div>
			</div>
		</div>

		<!-- Message Input -->
		<div class="d-flex align-items-end pt-2">
			<input
				type="file"
				ref="photoInput"
				class="d-none"
				@change="handlePhotoUpload"
			/>
			<button
				class="btn btn-outline-secondary me-2 position-relative"
				@click="triggerFileUpload"
				data-bs-toggle="tooltip"
				title="Attach a photo"
			>
				<img :src="ImageIcon" alt="Select Photo" width="24" />
			</button>

			<div
				class="flex-grow-1 d-flex align-items-center border rounded bg-light"
				style="height: 100%"
			>
				<!-- Photo Preview -->
				<div
					v-if="selectedPhoto"
					@keyup.enter="handleEnterPress"
					ref="photoPreviewDiv"
					tabindex="0"
					class="d-flex align-items-center w-100"
				>
					<div class="position-relative">
						<!-- Preview Image -->
						<img
							:src="photoPreview"
							alt="Photo Preview"
							class="rounded"
							style="
								width: auto;
								height: 100px;
								object-fit: cover;
								border: 1px solid #ddd;
							"
						/>

						<!-- Remove Photo Button -->
						<button
							class="btn btn-sm btn-danger rounded-circle position-absolute"
							style="
								top: -5px;
								right: -5px;
								width: 20px;
								height: 20px;
								font-size: 12px;
								padding: 0;
							"
							@click="selectedPhoto = null"
						>
							×
						</button>
					</div>
				</div>

				<input
					v-else
					v-model="newMessage"
					type="text"
					class="form-control border-0 bg-transparent w-100"
					placeholder="Type a message..."
					@keyup.enter="handleEnterPress"
				/>
			</div>
			<button class="btn btn-primary ms-2" @click="sendMessage">
				<img :src="SendIcon" alt="Send Message" width="24" />
			</button>
		</div>
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
						@click="handleUpdateGroupName"
						:disabled="updatingName"
					>
						{{ updatingName ? "Updating..." : "Update" }}
					</button>
				</div>
			</div>
		</div>
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
