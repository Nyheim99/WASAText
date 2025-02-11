<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";
import { ref, onMounted, onUnmounted } from "vue";
import axios from "../services/axios";

export default {
	props: {
		user: {
			type: Object,
			required: true,
		},
		conversations: {
			type: Array,
			required: true,
		},
		allUsers: {
			type: Array,
			required: true,
		},
		selectedConversation: {
			type: Object,
			required: false,
			default: () => ({}),
		},
	},
	emits: ["feedback", "select-conversation", "conversation-created"],
	setup(props, { emit }) {
		const modalMode = ref("");
		const searchQuery = ref("");
		const searchResults = ref([]);

		const selectedUser = ref(null);
		const privateMessage = ref("");

		const showValidation = ref(false);
		const validationMessage = ref("");

		const selectedUsers = ref(new Set());
		const groupName = ref("");
		const groupPhoto = ref(null);
		const groupMessage = ref("");

		const formatTimestamp = (timestamp) => {
			const utcTimestamp = new Date(timestamp);
			const localTimestamp = new Date(
				utcTimestamp.getTime() -
					utcTimestamp.getTimezoneOffset() * 60000
			);

			const now = new Date();
			const diffInMs = now - localTimestamp;

			const seconds = diffInMs / 1000;
			const minutes = seconds / 60;
			const hours = minutes / 60;
			const days = hours / 24;
			const weeks = days / 7;
			const months = days / 30;
			const years = days / 365;

			if (hours < 12) {
				return localTimestamp.toLocaleTimeString("en-GB", {
					hour: "2-digit",
					minute: "2-digit",
					hour12: false,
				});
			} else if (hours < 24) {
				return `${Math.floor(hours)}H`;
			} else if (days < 7) {
				return `${Math.floor(days)}d`;
			} else if (weeks < 4) {
				return `${Math.floor(weeks)}w`;
			} else if (months < 12) {
				return `${Math.floor(months)}m`;
			} else {
				return `${Math.floor(years)}y`;
			}
		};

		const resetModalState = () => {
			showValidation.value = false;
			validationMessage.value = "";
			selectedUser.value = null;
			privateMessage.value = "";
			searchQuery.value = "";
			searchResults.value = [];
			selectedUsers.value.clear();
			groupName.value = "";
			groupPhoto.value = null;
			groupMessage.value = "";
		};

		const toggleUserSelection = (user) => {
			if (selectedUsers.value.has(user)) {
				selectedUsers.value.delete(user);
			} else {
				selectedUsers.value.add(user);
			}
		};

		const isUserSelected = (user) => selectedUsers.value.has(user);

		const setMode = (mode) => {
			modalMode.value = mode;
		};

		const searchUsers = () => {
			if (!searchQuery.value) {
				searchResults.value = [];
				return;
			}

			const privateConversationUsers = new Set(
				props.conversations
					.filter((conv) => conv.conversation_type === "private")
					.map((conv) => conv.display_name)
			);

			const availableUsers = props.allUsers.filter(
				(user) =>
					user.id !== props.user.id &&
					!privateConversationUsers.has(user.username) &&
					user.username
						.toLowerCase()
						.includes(searchQuery.value.toLowerCase())
			);

			if (availableUsers.length === 0) {
				searchResults.value = [];
				validationMessage.value =
					"No matching users found or already in a conversation.";
				showValidation.value = true;
			} else {
				searchResults.value = availableUsers;
				showValidation.value = false;
			}
		};

		const showAllUsersOnFocus = () => {
			const privateConversationUsers = new Set(
				props.conversations
					.filter((conv) => conv.conversation_type === "private")
					.map((conv) => conv.display_name)
			);

			const availableUsers = props.allUsers.filter(
				(user) =>
					user.id !== props.user.id &&
					!privateConversationUsers.has(user.username)
			);

			if (availableUsers.length === 0) {
				searchResults.value = [];
				validationMessage.value =
					props.allUsers.length <= 1
						? "No users available to start a conversation."
						: "You've already started private conversations with all available users.";
				showValidation.value = true;
			} else {
				searchResults.value = availableUsers;
				showValidation.value = false;
			}
		};

		const selectUser = (user) => {
			selectedUser.value = user;
			privateMessage.value = "";
			searchResults.value = [];
		};

		const createPrivateConversation = async () => {
			if (!selectedUser.value) {
				showValidation.value = true;
				validationMessage.value =
					"Please select a user to start a conversation.";
				return;
			}

			if (!validateMessage(privateMessage.value)) {
				showValidation.value = true;
				return;
			}

			const formData = new FormData();
			formData.append("conversation_type", "private");
			formData.append("message", privateMessage.value);
			formData.append("recipientID", selectedUser.value.id);

			try {
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				resetModalState();

				emit("conversation-created");
				emit("feedback", "Conversation started successfully!");

				const modal = document.getElementById("newConversationModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();
			} catch (error) {
				showValidation.value = true;
				validationMessage.value = "Failed to start conversation:";
				const errorMessage =
					error.response?.data?.message ||
					"Failed to start conversation.";
				emit("feedback", errorMessage);
			}
		};

		const createGroupConversation = async () => {
			if (!validateGroupName(groupName.value)) {
				showValidation.value = true;
				return;
			}

			if (selectedUsers.value.size === 0) {
				showValidation.value = true;
				validationMessage.value = "Please select at least one user.";
				return;
			}

			if (!validateMessage(groupMessage.value)) {
				showValidation.value = true;
				return;
			}

			const formData = new FormData();
			formData.append("conversation_type", "group");
			formData.append("group_name", groupName.value);
			formData.append("message", groupMessage.value);
			[...selectedUsers.value].forEach((user) =>
				formData.append("participants", user.id)
			);

			if (groupPhoto.value) {
				formData.append("group_photo", groupPhoto.value);
			}

			try {
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				emit("conversation-created");
				emit("feedback", "Group conversation created successfully!");

				const modal = document.getElementById("newConversationModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();
			} catch (error) {
				showValidation.value = true;
				validationMessage.value = "Failed to start conversation:";
				const errorMessage =
					error.response?.data?.message ||
					"Failed to start conversation.";
				emit("feedback", errorMessage);
			}
		};

		const validateGroupName = (groupname) => {
			if (groupname.length < 3 || groupname.length > 20) {
				validationMessage.value =
					"Group name must be between 3 and 20 characters long.";
				return false;
			}
			if (!/^[a-zA-Z0-9 ]*$/.test(groupname)) {
				validationMessage.value =
					"Group name can only contain alphanumeric characters.";
				return false;
			}
			return true;
		};

		const validateMessage = (message) => {
			if (message.trim().length < 1 || message.length > 1000) {
				validationMessage.value =
					"Message must be between 1 and 1000 characters long.";
				return false;
			}
			if (!/^[a-zA-Z0-9À-ÿ.,!?()\-\"' ]+$/.test(message)) {
				validationMessage.value =
					"Message contains invalid characters.";
				return false;
			}
			return true;
		};

		const resolvePhotoURL = (photoURL, conversationType) => {
			if (photoURL && photoURL.startsWith("/")) {
				return `${__API_URL__}${photoURL}`;
			}
			return conversationType === "group" ? PeopleIcon : AvatarIcon;
		};

		const truncateMessage = (message, maxLength) => {
			if (message.length > maxLength) {
				return message.slice(0, maxLength) + "...";
			}
			return message;
		};

		onMounted(() => {
			const modal = document.getElementById("newConversationModal");
			if (modal) {
				modal.addEventListener("hide.bs.modal", resetModalState);
			}
		});

		onUnmounted(() => {
			const modal = document.getElementById("newConversationModal");
			if (modal) {
				modal.removeEventListener("hide.bs.modal", resetModalState);
			}
		});

		return {
			AvatarIcon,
			searchQuery,
			searchResults,
			modalMode,
			setMode,
			searchUsers,
			showAllUsersOnFocus,
			selectUser,
			selectedUser,
			selectedUsers,
			createPrivateConversation,
			createGroupConversation,
			privateMessage,
			groupMessage,
			groupName,
			groupPhoto,
			toggleUserSelection,
			isUserSelected,
			formatTimestamp,
			resolvePhotoURL,
			truncateMessage,
			showValidation,
			validationMessage,
		};
	},
};
</script>

<template>
	<div
		class="rounded shadow-sm p-1 d-flex flex-column h-100"
		style="width: 300px; background-color: #c8e1ff"
	>
		<div class="container row w-100 m-0 p-0">
			<div class="d-flex justify-content-between align-items-center p-1">
				<h5 class="mb-0"><b>Conversations</b></h5>
				<button
					class="btn rounded-circle d-flex align-items-center justify-content-center"
					type="button"
					data-bs-toggle="modal"
					data-bs-target="#newConversationModal"
					title="New Conversation"
					style="
						width: 24px;
						height: 24px;
						background-color: #dfecff;
						color: #2c3e50;
						border: none;
					"
					onmouseover="this.style.backgroundColor='#B0CFFB';"
					onmouseout="this.style.backgroundColor='#DFECFF';"
				>
					<i class="bi bi-pencil-square"></i>
				</button>
			</div>
		</div>

		<div class="overflow-auto">
			<div
				v-if="conversations.length === 0"
				class="container d-flex flex-column align-items-center justify-content-center text-center p-4"
			>
				<p class="text-muted mb-2">You have no chats yet!</p>
				<button
					class="btn btn-secondary btn-sm d-flex align-items-center"
					data-bs-toggle="modal"
					data-bs-target="#newConversationModal"
				>
					<span class="me-2">Start a Conversation</span>
					<i class="bi bi-pencil-square"></i>
				</button>
			</div>

			<div
				v-for="conversation in conversations"
				v-else
				:key="conversation.conversation_id"
				class="container d-flex align-items-center p-2 border-bottom"
				:class="{
					'bg-light':
						selectedConversation?.conversation_id ===
						conversation.conversation_id,
				}"
				@click="$emit('select-conversation', conversation)"
			>
				<img
					:src="
						resolvePhotoURL(
							conversation.display_photo_url,
							conversation.conversation_type
						)
					"
					alt="Avatar"
					class="rounded-circle me-2 bg-white"
					style="width: 40px; height: 40px; object-fit: cover"
				/>

				<div class="flex-grow-1">
					<h6 class="mb-1">{{ conversation.display_name }}</h6>
					<span
						v-if="conversation.last_message_is_deleted"
						class="text-muted"
						style="font-size: 0.8rem"
					>
						<i
							>{{
								conversation.last_message_sender_id === user.id
									? "You"
									: conversation.last_message_sender
							}}
							deleted a message</i
						>
					</span>
					<p v-else class="mb-0 text-muted" style="font-size: 0.8rem">
						<strong
							v-if="conversation.conversation_type === 'group'"
						>
							{{ conversation.last_message_sender }}:
						</strong>
						<strong
							v-else-if="
								conversation.last_message_sender_id === user.id
							"
						>
							You:
						</strong>
						<span v-if="conversation.last_message_content">
							{{
								truncateMessage(
									conversation.last_message_content,
									20
								)
							}}
						</span>
						<svg
							v-else
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							fill="currentColor"
							class="bi bi-camera-fill"
							viewBox="0 0 16 16"
						>
							<path
								d="M10.5 8.5a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0"
							/>
							<path
								d="M2 4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-1.172a2 2 0 0 1-1.414-.586l-.828-.828A2 2 0 0 0 9.172 2H6.828a2 2 0 0 0-1.414.586l-.828.828A2 2 0 0 1 3.172 4zm.5 2a.5.5 0 1 1 0-1 .5.5 0 0 1 0 1m9 2.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0"
							/>
						</svg>
					</p>
				</div>

				<small
					class="text-muted"
					style="align-self: flex-start; font-size: 0.7rem"
				>
					{{ formatTimestamp(conversation.last_message_timestamp) }}
				</small>
			</div>
		</div>

		<div
			id="newConversationModal"
			class="modal fade"
			tabindex="-1"
			aria-labelledby="newConversationModalLabel"
			aria-hidden="true"
		>
			<div class="modal-dialog">
				<div class="modal-content">
					<div class="modal-header">
						<h5 id="newConversationModalLabel" class="modal-title">
							Start a New Conversation
						</h5>
						<button
							type="button"
							class="btn-close"
							data-bs-dismiss="modal"
							aria-label="Close"
						></button>
					</div>
					<div class="modal-body">
						<div class="mb-2">
							<strong>Step 1:</strong> Select conversation type
						</div>

						<div
							class="btn-group mb-3"
							role="group"
							aria-label="Conversation Type Selector"
						>
							<button
								type="button"
								class="btn"
								:class="
									modalMode === 'private'
										? 'btn-primary'
										: 'btn-light'
								"
								@click="setMode('private')"
							>
								Private
							</button>
							<button
								type="button"
								class="btn"
								:class="
									modalMode === 'group'
										? 'btn-primary'
										: 'btn-light'
								"
								@click="setMode('group')"
							>
								Group
							</button>
						</div>

						<div v-if="modalMode === 'private'">
							<div class="mb-2">
								<strong>Step 2:</strong> Select a user
							</div>
							<input
								v-model="searchQuery"
								type="text"
								class="form-control"
								placeholder="Search for a user"
								@input="searchUsers"
								@focus="showAllUsersOnFocus"
							/>
							<div
								v-if="searchResults.length > 0"
								class="position-relative"
							>
								<ul
									class="list-group position-absolute"
									style="
										z-index: 1050;
										top: 100%;
										left: 0;
										width: 100%;
										max-height: 200px;
										overflow-y: auto;
									"
								>
									<li
										v-for="user in searchResults"
										:key="user.id"
										class="list-group-item d-flex align-items-center bg-body-secondary"
										@click="selectUser(user)"
									>
										<img
											v-if="
												resolvePhotoURL(user.photo_url)
											"
											:src="
												resolvePhotoURL(user.photo_url)
											"
											alt="Avatar"
											class="rounded-circle me-2"
											style="
												width: 30px;
												height: 30px;
												object-fit: cover;
											"
										/>
										<img
											v-else
											:src="AvatarIcon"
											alt="Default Avatar"
											class="rounded-circle me-2"
											style="width: 30px; height: 30px"
										/>
										{{ user.username }}
									</li>
								</ul>
							</div>
							<div v-if="selectedUser" class="mt-3">
								<div class="mb-2">
									<strong>Step 3:</strong> Send the first
									message to {{ selectedUser.username }}!
								</div>
								<textarea
									v-model="privateMessage"
									class="form-control"
									placeholder="Write your message here..."
									rows="3"
								></textarea>
							</div>
							<p
								v-if="showValidation"
								class="text-danger small mt-2"
								aria-live="assertive"
							>
								{{ validationMessage }}
							</p>
						</div>

						<div v-if="modalMode === 'group'">
							<div class="mb-2">
								<strong>Step 2:</strong> Enter group name
							</div>
							<input
								v-model="groupName"
								type="text"
								class="form-control mb-2"
								placeholder="Group name..."
							/>
							<div v-if="groupName">
								<div class="mb-2">
									<strong>Step 3 (optional):</strong> Upload
									Group Photo
								</div>
								<input
									type="file"
									class="form-control mb-2"
									accept="image/jpeg, image/png"
									@change="
										(e) => (groupPhoto = e.target.files[0])
									"
								/>

								<div class="mb-2">
									<strong>Step 4:</strong> Select Group
									Members
								</div>

								<div class="dropdown">
									<button
										id="selectUsersDropdown"
										class="btn btn-light dropdown-toggle"
										type="button"
										data-bs-toggle="dropdown"
										aria-expanded="false"
									>
										Select Users
									</button>
									<ul
										class="dropdown-menu"
										aria-labelledby="selectUsersDropdown"
										style="
											max-height: 300px;
											overflow-y: auto;
										"
									>
										<li
											v-for="user in allUsers.filter(
												(u) => u.id !== user.id
											)"
											:key="user.id"
											class="dropdown-item d-flex align-items-center justify-content-between"
										>
											<div
												class="d-flex align-items-center"
											>
												<img
													:src="
														resolvePhotoURL(
															user.photo_url
														)
													"
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
												:checked="isUserSelected(user)"
												@click="
													toggleUserSelection(user)
												"
											/>
										</li>
									</ul>
								</div>

								<div v-if="selectedUsers.size > 0" class="mt-3">
									<h6>Users selected:</h6>
									<div class="d-flex flex-wrap gap-2">
										<span
											v-for="user in [...selectedUsers]"
											:key="user.id"
											class="badge text-bg-secondary d-flex align-items-center"
											style="
												font-size: 14px;
												padding: 0.5em 0.75em;
											"
										>
											{{ user.username }}
											<button
												class="btn-close btn-close-white ms-2"
												aria-label="Remove"
												style="
													font-size: 10px;
													opacity: 0.8;
												"
												@click="
													toggleUserSelection(user)
												"
											></button>
										</span>
									</div>
								</div>

								<div class="my-2">
									<strong>Step 5:</strong> Write the first
									message!
								</div>

								<textarea
									v-model="groupMessage"
									class="form-control mt-2"
									placeholder="Write an initial message..."
									rows="3"
								></textarea>
							</div>

							<p
								v-if="showValidation"
								class="text-danger small mt-2"
								aria-live="assertive"
							>
								{{ validationMessage }}
							</p>
						</div>
					</div>
					<div class="modal-footer">
						<div
							v-if="modalMode === 'private'"
							class="mt-2 d-flex justify-content-end"
						>
							<button
								type="button"
								class="btn btn-primary"
								@click="createPrivateConversation"
							>
								Start Conversation!
							</button>
						</div>
						<div v-else class="mt-2 d-flex justify-content-end">
							<button
								type="button"
								class="btn btn-primary"
								@click="createGroupConversation"
							>
								Start Conversation!
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
