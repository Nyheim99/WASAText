<script>
import WriteIcon from "../assets/pencil-square.svg";
import { ref, onMounted } from "vue";
import axios from "../services/axios";

export default {
	props: {
		conversations: {
			type: Array,
			required: true,
			default: () => [],
		},
	},
	setup() {
		const modalMode = ref("private");
		const searchQuery = ref("");
		const searchResults = ref([]);

		const selectedUser = ref(null);
		const message = ref("");

		const allUsers = ref([]);

		const fetchUsers = async () => {
			try {
				const response = await axios.get("/users");
				allUsers.value = response.data;
			} catch (error) {
				console.error("Failed to fetch users:", error.message);
			}
		};

		const setMode = (mode) => {
			modalMode.value = mode;
		};

		const searchUsers = () => {
			if (!searchQuery.value) {
				searchResults.value = [];
				return;
			}
			searchResults.value = allUsers.value
				.filter((user) =>
					user.username
						.toLowerCase()
						.includes(searchQuery.value.toLowerCase())
				)
				.map((user) => user.username);
		};

		const showAllUsersOnFocus = () => {
			searchResults.value = allUsers.value.map((user) => user.username);
		};

		const selectUser = (username) => {
			selectedUser.value = username;
			message.value = "";
			searchResults.value = [];
		};

		const createPrivateConversation = async () => {
			if (!message.value.trim()) {
				alert("Message cannot be empty.");
				return;
			}

			if (!selectedUser.value) {
				alert("Please select a user to start a conversation.");
				return;
			}

			try {
				await axios.post("/conversations", {
					conversationType: "private",
					message: message.value,
					username: selectedUser.value,
					timestamp: new Date().toISOString(),
				});

				const newConversation = response.data;

				conversations.push(newConversation);

				selectedUser.value = null;
				message.value = "";
				searchQuery.value = "";
				searchResults.value = [];
			} catch (error) {
				console.error("Failed to start conversation:", error.message);
				alert("Failed to start conversation.");
			}
		};

		onMounted(() => {
			fetchUsers();
		});

		return {
			WriteIcon,
			searchQuery,
			searchResults,
			modalMode,
			setMode,
			searchUsers,
			showAllUsersOnFocus,
			selectUser,
			selectedUser,
			createPrivateConversation,
			message,
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
							aria-label="Basic example"
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

						<input
							v-if="modalMode === 'private'"
							type="text"
							class="form-control"
							placeholder="Search for a user"
							v-model="searchQuery"
							@input="searchUsers"
							@focus="showAllUsersOnFocus"
						/>
						<div class="position-relative">
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
									:key="user"
									class="list-group-item bg-body-secondary"
									@click="selectUser(user)"
								>
									{{ user }}
								</li>
							</ul>
						</div>

						<div v-if="selectedUser" class="mt-3">
							<h6>Send a message to {{ selectedUser }}</h6>
							<textarea
								v-model="message"
								class="form-control"
								placeholder="Write your message here..."
								rows="3"
							></textarea>
							<div class="mt-2 d-flex justify-content-end">
								<button
									type="button"
									class="btn btn-secondary me-2"
									@click="selectedUser = null"
								>
									Back
								</button>
								<button
									type="button"
									class="btn btn-primary"
									@click="createPrivateConversation"
								>
									Send Message
								</button>
							</div>
						</div>

						<!-- Group Conversation -->
						<div v-if="modalMode === 'group'" class="mt-3">
							<ul class="list-group">
								<li
									v-for="user in searchResults"
									:key="user"
									class="list-group-item"
								>
									{{ user }}
								</li>
							</ul>

							<div class="mt-3">
								<h6>Selected Users:</h6>
								<ul class="list-group">
									<li
										v-for="user in selectedUsers"
										:key="user"
										class="list-group-item d-flex justify-content-between"
									>
										{{ user }}
										<button
											type="button"
											class="btn btn-sm btn-danger"
										>
											Remove
										</button>
									</li>
								</ul>
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
						<button type="button" class="btn btn-primary">
							Send message
						</button>
					</div>
				</div>
			</div>
		</div>

		<ul v-if="conversations.length > 0" class="list-unstyled">
			<li
				v-for="conversation in conversations"
				:key="conversation.conversationId"
				class="p-2 mb-2 bg-secondary text-white rounded"
			>
				{{ conversation.name }}
			</li>
		</ul>
		<p v-else class="text-muted text-center">
			No conversations found. Click the button above to start a new
			conversation!
		</p>
	</div>
</template>
