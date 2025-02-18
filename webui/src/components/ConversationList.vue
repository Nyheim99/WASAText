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

		//Helper function to format the preview timestamp
		const formatTimestamp = (timestamp) => {
			const utcTimestamp = new Date(timestamp);
			const localTimestamp = new Date(
				utcTimestamp.getTime() -
					utcTimestamp.getTimezoneOffset() * 60000
			);

			return localTimestamp.toLocaleTimeString("en-GB", {
				day: "2-digit",
				month: "2-digit",
				hour: "2-digit",
				minute: "2-digit",
				hour12: false,
			});
		};

		//Helper funciton to reset the modal states
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

		//User selection when creating group conversations
		const toggleUserSelection = (user) => {
			if (selectedUsers.value.has(user)) {
				selectedUsers.value.delete(user);
			} else {
				selectedUsers.value.add(user);
			}
		};

		//Check if user is selected
		const isUserSelected = (user) => selectedUsers.value.has(user);

		//Set modal mode
		const setMode = (mode) => {
			modalMode.value = mode;
		};

		//Search for users when creating conversation
		const searchUsers = () => {
			if (!searchQuery.value) {
				searchResults.value = [];
				return;
			}

			//Create a set of all users
			const privateConversationUsers = new Set(
				props.conversations
					.filter((conv) => conv.conversation_type === "private")
					.map((conv) => conv.display_name)
			);

			//Only display users the user does not already have a conversation with
			const availableUsers = props.allUsers.filter(
				(user) =>
					user.id !== props.user.id &&
					!privateConversationUsers.has(user.username) &&
					user.username
						.toLowerCase()
						.includes(searchQuery.value.toLowerCase())
			);

			//Deal with no available users
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

		//Show all available users when the search bar is focused
		const showAllUsersOnFocus = () => {
			const privateConversationUsers = new Set(
				props.conversations
					.filter((conv) => conv.conversation_type === "private")
					.map((conv) => conv.display_name)
			);

			//Filter out users the that the user already has a conversation with
			const availableUsers = props.allUsers.filter(
				(user) =>
					user.id !== props.user.id &&
					!privateConversationUsers.has(user.username)
			);

			//Deal with no users left
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

		//Selecting a user to start a private conversation
		const selectUser = (user) => {
			selectedUser.value = user;
			privateMessage.value = "";
			searchResults.value = [];
		};

		//Creating a private conversation
		const createPrivateConversation = async () => {
			//Validate all the input
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

			//Append all the data
			const formData = new FormData();
			formData.append("conversation_type", "private");
			formData.append("message", privateMessage.value);
			formData.append("recipientID", selectedUser.value.id);

			try {
				//Attempt to create the conversation
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				//Resets the modal state
				resetModalState();

				//Emits the feedback to HomeView and tells it to update conversations
				emit("conversation-created");
				emit("feedback", "Conversation started successfully!");

				//Close the modal on success
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

		//Creating a group conversation
		const createGroupConversation = async () => {
			//Validate the inputs
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

			//Append all the data
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
				//Attempt to create the conversation
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				//Notify HomeView of the new conversation
				emit("conversation-created");
				emit("feedback", "Group conversation created successfully!");

				//Close the modal on success
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

		//Validate the new group name
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

		//Validate the sent message
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

		//Get the picture source for display photo
		const resolvePhotoURL = (photoURL, conversationType) => {
			if (photoURL && photoURL.startsWith("/")) {
				return `${__API_URL__}${photoURL}`;
			}
			return conversationType === "group" ? PeopleIcon : AvatarIcon;
		};

		//Truncate long messages
		const truncateMessage = (message, maxLength) => {
			if (message.length > maxLength) {
				return message.slice(0, maxLength) + "...";
			}
			return message;
		};

		//Functionality to reset modal states on component loads
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
		class="bg-white rounded shadow-sm p-1 d-flex flex-column h-100"
		style="width: 300px"
	>
		<!-- Conversation List -->
		<div class="container row w-100 m-0 p-0">

			<!-- Top Section -->
			<div class="d-flex justify-content-between align-items-center p-1">
				<h5 class="mb-2"><b>Conversations</b></h5>
				<button
					class="btn bg-body-secondary rounded-circle d-flex align-items-center justify-content-center hover-bg"
					type="button"
					data-bs-toggle="modal"
					data-bs-target="#newConversationModal"
					title="New Conversation"
					style="width: 24px; height: 24px; border: none"
				>
					<i class="bi bi-pencil-square"></i>
				</button>
			</div>
		</div>

		<!-- Conversation Previews-->
		<div class="overflow-auto">

			<!-- If user has no conversations-->
			<div
				v-if="conversations.length === 0"
				class="container d-flex flex-column align-items-center justify-content-center text-center p-4"
			>
				<p class="text-muted mb-2">You have no chats yet!</p>
				<button
					class="btn btn-sm d-flex align-items-center"
					data-bs-toggle="modal"
					data-bs-target="#newConversationModal"
					style="background-color: #dfecff"
				>
					<span class="me-2">Start a Conversation</span>
					<i class="bi bi-pencil-square"></i>
				</button>
			</div>

			<!-- Display conversations in a list-->
			<div
				v-for="conversation in conversations"
				v-else
				:key="conversation.conversation_id"
				class="container d-flex align-items-center p-2 border-bottom hover-bg"
				:class="
					selectedConversation?.conversation_id ===
						conversation.conversation_id && 'bg-body-secondary'
				"
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
					style="
						width: 40px;
						height: 40px;
						object-fit: cover;
						min-width: 40px;
					"
				/>

				<!-- Single Conversation Preview-->
				<div class="flex-grow-1 overflow-hidden">
					<h6 class="mb-1">{{ conversation.display_name }}</h6>

					<!-- Deleted messages -->
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

					<!-- Normal messages -->
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
						<i v-else class="bi bi-image-fill"></i>
					</p>
				</div>

				<!-- Timestamp -->
				<small
					class="text-muted p-1"
					style="align-self: flex-start; font-size: 0.7rem"
				>
					{{ formatTimestamp(conversation.last_message_timestamp) }}
				</small>
			</div>
		</div>

		<!-- Modal for creating new conversations -->
		<div
			id="newConversationModal"
			class="modal fade"
			tabindex="-1"
			aria-labelledby="newConversationModalLabel"
			aria-hidden="true"
		>
			<div class="modal-dialog">
				<div class="modal-content">

					<!-- Header section -->
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

					<!-- Group Creation Form -->
					<div class="modal-body">

						<!-- Conversation Type selector -->
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

						<!-- Form for creating private conversation-->
						<div v-if="modalMode === 'private'">

							<!-- Selecting recipient with search-->
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

							<!-- Initial private message input -->
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

							<!-- Validation field for errors-->
							<p
								v-if="showValidation"
								class="text-danger small mt-2"
								aria-live="assertive"
							>
								{{ validationMessage }}
							</p>
						</div>

						<!-- Form for creating private conversation-->
						<div v-if="modalMode === 'group'">

							<!-- Group name input -->
							<div class="mb-2">
								<strong>Step 2:</strong> Enter group name
							</div>
							<input
								v-model="groupName"
								type="text"
								class="form-control mb-2"
								placeholder="Group name..."
							/>

							<!-- Optional group photo input -->
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

								<!-- Group member selection -->
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

								<!-- Selected users display -->
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

								<!-- Initial group message input -->
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

							<!-- Validation field for errors -->
							<p
								v-if="showValidation"
								class="text-danger small mt-2"
								aria-live="assertive"
							>
								{{ validationMessage }}
							</p>
						</div>
					</div>

					<!-- Modal footer-->
					<div class="modal-footer">

						<!-- Button for creating private conversation -->
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

						<!-- Button for creating group conversation -->
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

<style scoped>
.hover-bg:hover {
	background-color: var(--bs-gray-100);
	transition: background-color 0.3s ease;
	cursor: pointer;
}
</style>
