<script>
import WriteIcon from "../assets/pencil-square.svg";
import AvatarIcon from "../assets/person-circle.svg";
import { ref, onMounted } from "vue";
import axios from "../services/axios";
const backendBaseURL = "http://localhost:3000";

export default {
	emits: ["feedback"],
	props: {
		user: {
			type: Object,
			required: true,
		},
	},
	setup(props, { emit }) {
		const modalMode = ref("private");
		const searchQuery = ref("");
		const searchResults = ref([]);
		const selectedUser = ref(null);
		const selectedUsers = ref(new Set());
		const privateMessage = ref("");
    const groupMessage = ref("");
		const allUsers = ref([]);
		const feedbackMessage = ref("");
		const showFeedback = ref(false);
		const groupName = ref("");
		const groupPhoto = ref(null);

		const fetchUsers = async () => {
			try {
				const response = await axios.get("/users");
				const filteredUsers = response.data.filter(
					(user) => user.id !== props.user.id
				);
				allUsers.value = filteredUsers;
			} catch (error) {
				console.error("Failed to fetch users:", error.message);
			}
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
			searchResults.value = allUsers.value.filter((user) =>
				user.username
					.toLowerCase()
					.includes(searchQuery.value.toLowerCase())
			);
		};

		const showAllUsersOnFocus = () => {
			searchResults.value = allUsers.value;
		};

		const selectUser = (user) => {
			selectedUser.value = user;
			privateMessage.value = "";
			searchResults.value = [];
		};

		const createPrivateConversation = async () => {
			if (!privateMessage.value.trim()) {
				alert("Message cannot be empty.");
				return;
			}

			if (!selectedUser.value) {
				alert("Please select a user to start a conversation.");
				return;
			}

			const formData = new FormData();
			formData.append("conversation_type", "private");
			formData.append("message", privateMessage.value);
			formData.append("username", selectedUser.value.username);

			try {
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				selectedUser.value = null;
				privateMessage.value = "";
				searchQuery.value = "";
				searchResults.value = [];

				const modal = document.getElementById("newConversationModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();

				emit("feedback", "Conversation started successfully!");
			} catch (error) {
				console.error("Failed to start conversation:", error.message);
				const errorMessage =
					error.response?.data?.message ||
					"Failed to start conversation.";
				emit("feedback", errorMessage);
			}
		};

		const createGroupConversation = async () => {
			if (!groupName.value.trim()) {
				alert("Group name is required.");
				return;
			}

			if (!groupMessage.value.trim()) {
				alert("Message cannot be empty.");
				return;
			}

			if (selectedUsers.value.size === 0) {
				alert("Please select at least one user.");
				return;
			}

			const formData = new FormData();
			formData.append("conversation_type", "group");
			formData.append("group_name", groupName.value);
			formData.append("message", groupMessage.value);
			formData.append(
				"participants",
				JSON.stringify([...selectedUsers.value].map((user) => user.id))
			);

			if (groupPhoto.value) {
				formData.append("group_photo", groupPhoto.value);
			}

			try {
				await axios.post("/conversations", formData, {
					headers: { "Content-Type": "multipart/form-data" },
				});

				selectedUsers.value.clear();
				groupName.value = "";
				groupPhoto.value = null;
				groupMessage.value = "";
				searchQuery.value = "";
				searchResults.value = [];

				const modal = document.getElementById("newConversationModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();

				emit("feedback", "Group conversation created successfully!");
			} catch (error) {
				console.error(
					"Failed to create group conversation:",
					error.message
				);
				const errorMessage =
					error.response?.data?.message ||
					"Failed to create group conversation.";
				emit("feedback", errorMessage);
			}
		};

		onMounted(() => {
			fetchUsers();
		});

		return {
			WriteIcon,
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
			feedbackMessage,
			showFeedback,
			allUsers,
			resolvePhotoURL: (photoURL) =>
				photoURL && photoURL.trim() !== ""
					? `${backendBaseURL}${photoURL}`
					: AvatarIcon,
		};
	},
};
</script>

<template>
	<div class="rounded shadow-sm bg-white p-3 overflow-auto h-100">
		<div class="container mb-3">
			<div class="row">
				<div class="d-flex justify-content-between align-items-center">
					<h5 class="mb-0" style="line-height: 1.5">Conversations</h5>
					<button
						class="btn btn-light p-1 d-flex align-items-center justify-content-center"
						type="button"
						data-bs-toggle="modal"
						data-bs-target="#newConversationModal"
						style="width: 24px; height: 24px"
					>
						<img :src="WriteIcon" alt="New Conversation" />
					</button>
				</div>
			</div>
		</div>

		<div
			class="modal fade"
			id="newConversationModal"
			tabindex="-1"
			aria-labelledby="newConversationModalLabel"
			aria-hidden="true"
		>
			<div class="modal-dialog">
				<div class="modal-content">
					<div class="modal-header">
						<h5 class="modal-title" id="newConversationModalLabel">
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
							<input
								type="text"
								class="form-control"
								placeholder="Search for a user"
								v-model="searchQuery"
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
							<h6>
								Send a message to {{ selectedUser.username }}
							</h6>
							<textarea
								v-model="privateMessage"
								class="form-control"
								placeholder="Write your message here..."
								rows="3"
							></textarea>
						</div>
						</div>

						

						<div v-if="modalMode === 'group'">
							<input
								type="text"
								class="form-control mb-2"
								placeholder="Enter group name"
								v-model="groupName"
							/>
							<input
								type="file"
								class="form-control mb-2"
								accept="image/*"
								@change="
									(e) => (groupPhoto = e.target.files[0])
								"
							/>

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
										v-for="user in allUsers"
										:key="user.id"
										class="dropdown-item d-flex align-items-center justify-content-between"
									>
										<div class="d-flex align-items-center">
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
											@click="toggleUserSelection(user)"
										/>
									</li>
								</ul>
							</div>

							<div class="mt-3">
								<h6>Selected Users:</h6>
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
											@click="toggleUserSelection(user)"
											style="
												font-size: 10px;
												opacity: 0.8;
											"
										></button>
									</span>
								</div>
							</div>

							<textarea
								v-model="groupMessage"
								class="form-control mt-2"
								placeholder="Write an initial message..."
								rows="3"
							></textarea>
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
								Send Message Priv
							</button>
						</div>
						<div v-else class="mt-2 d-flex justify-content-end">
							<button
								type="button"
								class="btn btn-primary"
								@click="createGroupConversation"
							>
								Send message Group
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<p class="text-muted text-center">
			No conversations found. Click the button above to start a new
			conversation!
		</p>
	</div>
</template>
