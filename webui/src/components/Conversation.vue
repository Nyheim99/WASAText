<script>
import AvatarIcon from "/person-fill.svg";
import PeopleIcon from "/people-fill.svg";
import ImageIcon from "/image.svg";
import SendIcon from "/send.svg";
import {
	ref,
	onMounted,
	onBeforeUnmount,
	watch,
	computed,
	nextTick,
} from "vue";
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
		"message-deleted",
		"message-forwarded",
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
		const messageToDelete = ref(null);
		const forwardMessage = ref(null);
		const replyToMessage = ref(null);
		const availableConversations = ref([]);
		const messages = computed(
			() => props.conversationDetails?.messages || []
		);
		const messageContainer = ref(null);

		const contextMenu = ref({ visible: false, x: 0, y: 0, message: null });

		const reactionPicker = ref({
			visible: false,
			x: 0,
			y: 0,
			message: null,
		});
		const availableReactions = ["👍", "😂", "❤️", "🔥", "😢"];
		let activePopover = null;

		const conversationPhoto = () => {
			if (props.conversation.display_photo_url?.startsWith("/")) {
				return `${__API_URL__}${props.conversation.display_photo_url}?t=${cacheBuster.value}`;
			}
			return props.conversation.conversation_type === "group"
				? PeopleIcon
				: AvatarIcon;
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
			if (userId === props.user.id) {
				return props.user.username;
			}

			const user = props.allUsers.find((user) => user.id === userId);
			return user ? user.username : "Unknown";
		};

		const handleUpdateGroupName = async () => {
			if (!validateGroupName(newGroupName.value.trim())) {
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

				props.conversation.display_name = response.data;

				emit("group-name-updated", {
					conversationId: props.conversation.conversation_id,
					newName: response.data,
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

				cacheBuster.value = Date.now();

				emit("group-photo-updated", {
					conversationId: props.conversation.conversation_id,
					newPhotoUrl:
						response.data.photo_url + "?t=" + cacheBuster.value,
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

		const validateImage = (file) => {
			const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
			if (!allowedTypes.includes(file.type)) {
				validationMessage.value =
					"Only JPEG and PNG files are allowed.";
				return false;
			}
			return true;
		};

		const handleAddMembers = async () => {
			if (selectedUsers.value.size === 0) {
				validationMessage.value = "Please select at least one user.";
				return;
			}

			addingMembers.value = true;

			try {
				await axios.post(
					`/conversations/${props.conversation.conversation_id}/members`,
					{
						participants: [...selectedUsers.value],
					}
				);

				emit(
					"group-members-updated",
					props.conversation.conversation_id
				);

				selectedUsers.value.clear();
				updateUsersNotInGroup();

				const modal = document.getElementById("addMembersModal");
				if (modal) {
					const bootstrapModal = bootstrap.Modal.getInstance(modal);
					if (bootstrapModal) {
						bootstrapModal.hide();
					}
				}
			} catch (error) {
				console.error("Failed to add members:", error);
			} finally {
				addingMembers.value = false;
			}
		};

		const handleLeaveGroup = async () => {
			try {
				await axios.delete(
					`/conversations/${props.conversation.conversation_id}/members/me`
				);

				emit("group-left", props.conversation.conversation_id);

				const modalElement = document.getElementById("leaveGroupModal");
				const modal = bootstrap.Modal.getInstance(modalElement);
				if (modal) modal.hide();

				props.conversationDetails = null;
			} catch (error) {
				console.error("Failed to leave group:", error);
			}
		};

		const sendMessage = async () => {
			if (!newMessage.value && !selectedPhoto.value) {
				return;
			}

			const formData = new FormData();
			if (selectedPhoto.value) {
				formData.append("photo", selectedPhoto.value);
			} else {
				formData.append("message", newMessage.value);
			}

			if (replyToMessage.value) {
				formData.append("original_message_id", replyToMessage.value.id);
			}

			try {
				await axios.post(
					`/conversations/${props.conversation.conversation_id}/messages`,
					formData,
					{ headers: { "Content-Type": "multipart/form-data" } }
				);

				newMessage.value = "";
				selectedPhoto.value = null;
				replyToMessage.value = null;

				emit("message-sent", props.conversation.conversation_id);
			} catch (error) {
				console.error(
					"Failed to send message:",
					error.response?.data || error.message
				);
			}
		};

		const confirmForwardMessage = (message) => {
			forwardMessage.value = message;
			fetchAvailableConversations();
			const modal = new bootstrap.Modal(
				document.getElementById("forwardMessageModal")
			);
			modal.show();
		};

		const fetchAvailableConversations = async () => {
			try {
				const response = await axios.get("/conversations");
				availableConversations.value = response.data.filter(
					(conv) =>
						conv.conversation_id !==
						props.conversation.conversation_id
				);
			} catch (error) {
				console.error("Failed to fetch conversations:", error);
			}
		};

		const forwardMessageTo = async (conversationId) => {
			try {
				await axios.post(
					`/conversations/${conversationId}/messages/${forwardMessage.value.id}/forward`
				);

				emit("message-forwarded", props.conversation.conversation_id);

				forwardMessage.value = null;
				const modal = bootstrap.Modal.getInstance(
					document.getElementById("forwardMessageModal")
				);
				if (modal) modal.hide();
			} catch (error) {
				console.error("Failed to forward message:", error);
			}
		};

		const deleteMessage = async () => {
			try {
				await axios.delete(
					`/conversations/${props.conversation.conversation_id}/messages/${messageToDelete.value.id}`
				);

				emit("message-deleted", props.conversation.conversation_id);

				messageToDelete.value = null;

				const deleteModal = bootstrap.Modal.getInstance(
					document.getElementById("deleteMessageModal")
				);
				if (deleteModal) deleteModal.hide();
			} catch (error) {
				console.error("Failed to delete message:", error);
			} finally {
				contextMenu.value.visible = false;
			}
		};

		const confirmDeleteMessage = (message) => {
			messageToDelete.value = message;
			const deleteModal = new bootstrap.Modal(
				document.getElementById("deleteMessageModal")
			);
			deleteModal.show();
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

		const scrollToBottom = () => {
			nextTick(() => {
				if (messageContainer.value) {
					messageContainer.value.scrollTop =
						messageContainer.value.scrollHeight;
				}
			});
		};

		watch(
			() => messages.value.length,
			() => {
				scrollToBottom();
			}
		);

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

			if (selectedPhoto.value || newMessage.value.trim()) {
				sendMessage();
			}
		};

		const showContextMenu = (event, message) => {
			if (message.is_deleted) {
				return;
			}
			event.preventDefault();
			event.stopPropagation();

			const menuWidth = 75;

			let x = event.clientX;
			let y = event.clientY;

			if (message.sender_id === props.user.id) {
				x -= menuWidth;
			}

			contextMenu.value = {
				visible: true,
				x,
				y,
				message,
				canDelete: message.sender_id === props.user.id,
			};
		};

		const showReactionPicker = (event, message) => {
			event.preventDefault();
			event.stopPropagation();

			const pickerWidth = 200;
			const pickerHeight = 50;

			let x = event.clientX;
			let y = event.clientY;

			const viewportWidth = window.innerWidth;
			const viewportHeight = window.innerHeight;

			if (x + pickerWidth > viewportWidth) {
				x = viewportWidth - pickerWidth - 15;
			}

			if (y + pickerHeight > viewportHeight) {
				y = viewportHeight - pickerHeight - 10;
			}

			reactionPicker.value = {
				visible: true,
				x,
				y,
				message,
			};

			contextMenu.value.visible = false;
		};

		const toggleReaction = async (message, emoji) => {
			try {
				const userReaction = message.reactions?.find(
					(r) => r.user_id === props.user.id
				);

				if (userReaction?.emoticon === emoji) {
					// If user already reacted with this emoji → remove reaction
					await axios.delete(
						`/conversations/${props.conversation.conversation_id}/messages/${message.id}/reactions/me`
					);
					message.reactions = message.reactions.filter(
						(r) => r.user_id !== props.user.id
					);
				} else {
					// Otherwise → add/update reaction
					await axios.post(
						`/conversations/${props.conversation.conversation_id}/messages/${message.id}/reactions`,
						{ emoticon: emoji }
					);
					// Update the local message object optimistically
					message.reactions = message.reactions.filter(
						(r) => r.user_id !== props.user.id
					);
					message.reactions.push({
						user_id: props.user.id,
						emoticon: emoji,
					});
				}
			} catch (error) {
				console.error("Error toggling reaction:", error);
			} finally {
				reactionPicker.value.visible = false;
			}
		};

		const groupReactions = (reactions) => {
			const grouped = {};
			reactions.forEach((reaction) => {
				if (grouped[reaction.emoticon]) {
					grouped[reaction.emoticon] += 1;
				} else {
					grouped[reaction.emoticon] = 1;
				}
			});
			return grouped;
		};

		const showReactionDetails = (event, message) => {
			if (activePopover) {
				activePopover.dispose();
				activePopover = null;
			}

			const reactionList = message.reactions
				.map(
					(r) => `
            <div style="display: flex; align-items: center; gap: 30px;">
                <span style="flex-grow: 1; text-align: left;">${getUsername(
					r.user_id
				)}</span>
                <span style="font-size: 20px;">${r.emoticon}</span>
            </div>
        `
				)
				.join("");

			const popoverContent =
				reactionList.length > 0
					? reactionList
					: "<div>No reactions yet</div>";

			const popover = new bootstrap.Popover(event.currentTarget, {
				content: popoverContent,
				html: true,
				trigger: "manual",
				placement: "top",
				sanitize: false,
			});

			popover.show();
			activePopover = popover;
		};

		document.addEventListener("click", () => {
			contextMenu.value.visible = false;
			reactionPicker.value.visible = false;
			if (
				activePopover &&
				!event.target.closest("[data-bs-toggle='popover']")
			) {
				activePopover.dispose();
				activePopover = null;
			}
		});

		const setReplyTo = (message) => {
			replyToMessage.value = message;
			scrollToBottom();
		};

		const cancelReply = () => {
			replyToMessage.value = null;
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
			handleEnterPress,
			contextMenu,
			showContextMenu,
			reactionPicker,
			availableReactions,
			showReactionPicker,
			toggleReaction,
			showReactionDetails,
			deleteMessage,
			confirmDeleteMessage,
			messageContainer,
			scrollToBottom,
			groupReactions,
			forwardMessageTo,
			confirmForwardMessage,
			availableConversations,
			setReplyTo,
			cancelReply,
			replyToMessage,
		};
	},
};
</script>

<template>
	<div class="bg-white d-flex flex-column shadow-sm rounded h-100">
		<div
			class="d-flex align-items-center mb-2 p-4 pb-0 justify-content-between"
		>
			<div class="d-flex align-items-center">
				<img
					:src="conversationPhoto()"
					alt="Conversation Avatar"
					class="rounded-circle"
					style="width: 50px; height: 50px; object-fit: cover"
				/>
				<h2 class="px-2 mb-0">
					{{ conversation.display_name }}
				</h2>
			</div>

			<div class="dropdown ms-2">
				<button
					class="btn btn-light p-1 d-flex align-items-center justify-content-center rounded-circle"
					type="button"
					id="groupActionsDropdown"
					data-bs-toggle="dropdown"
					aria-expanded="false"
					data-bs-placement="top"
					title="Group Actions"
					style="width: 36px; height: 36px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);"
				>
					<i class="bi bi-three-dots fs-4"></i>
				</button>
				<ul
					class="dropdown-menu"
					aria-labelledby="groupActionsDropdown"
				>
					<li>
						<button
							class="dropdown-item"
							data-bs-toggle="modal"
							data-bs-target="#groupNameModal"
						>
							Update Group Name
						</button>
					</li>
					<li>
						<button
							class="dropdown-item"
							data-bs-toggle="modal"
							data-bs-target="#groupPhotoModal"
						>
							Update Group Picture
						</button>
					</li>
					<li>
						<button
							class="dropdown-item"
							data-bs-toggle="modal"
							data-bs-target="#addMembersModal"
						>
							Add Members
						</button>
					</li>
					<li>
						<hr class="dropdown-divider" />
					</li>
					<li>
						<button
							class="dropdown-item text-danger"
							data-bs-toggle="modal"
							data-bs-target="#leaveGroupModal"
						>
							Leave Group
						</button>
					</li>
				</ul>
			</div>
		</div>

		<hr class="m-0" />

		<!-- Context Menu -->
		<div
			v-if="contextMenu.visible"
			class="position-absolute bg-white border rounded shadow-sm p-1"
			:style="{
				zIndex: '1000',
				top: contextMenu.y + 'px',
				left: contextMenu.x + 'px',
				width: '150px',
			}"
		>
			<div
				class="px-3 py-2 rounded hover-bg"
				style="cursor: pointer"
				@click="showReactionPicker($event, contextMenu.message)"
			>
				React
			</div>
			<div
				class="px-3 py-2 rounded hover-bg"
				style="cursor: pointer"
				@click="confirmForwardMessage(contextMenu.message)"
			>
				Forward
			</div>
			<div
				class="px-3 py-2 rounded hover-bg"
				style="cursor: pointer"
				@click="setReplyTo(contextMenu.message)"
			>
				Reply
			</div>
			<hr class="my-1" v-if="contextMenu.canDelete" />
			<div
				v-if="contextMenu.canDelete"
				class="px-3 py-2 text-danger rounded hover-bg"
				style="cursor: pointer"
				@click="confirmDeleteMessage(contextMenu.message)"
			>
				Delete
			</div>
		</div>

		<!-- Reaction Picker -->
		<div
			v-if="reactionPicker.visible"
			:style="{
				position: 'absolute',
				background: 'white',
				border: '1px solid #ccc',
				padding: '5px 10px',
				borderRadius: '5px',
				boxShadow: '0px 2px 5px rgba(0, 0, 0, 0.2)',
				cursor: 'pointer',
				zIndex: '1000',
				top: reactionPicker.y + 'px',
				left: reactionPicker.x + 'px',
				display: 'flex',
				gap: '8px',
			}"
		>
			<span
				v-for="emoji in availableReactions"
				:key="emoji"
				@click="toggleReaction(reactionPicker.message, emoji)"
				:style="{
					fontSize: '20px',
					cursor: 'pointer',
					width: '28px',
					height: '28px',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					borderRadius: '50%',
					backgroundColor: reactionPicker.message?.reactions?.some(
						(r) => r.user_id === user.id && r.emoticon === emoji
					)
						? '#ddd'
						: 'transparent',
				}"
			>
				{{ emoji }}
			</span>
		</div>

		<div
			class="flex-grow-1 overflow-auto p-2 d-flex flex-column"
			ref="messageContainer"
		>
			<div
				v-for="message in messages"
				:key="message.id"
				class="d-flex align-items-start"
				:style="{
					justifyContent:
						message.sender_id === user.id
							? 'flex-end'
							: 'flex-start',
					marginBottom: message.reactions.length > 0 ? '26px' : '5px',
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
					class="p-2 rounded shadow-sm position-relative"
					@contextmenu.prevent="showContextMenu($event, message)"
					:style="{
						backgroundColor:
							message.sender_id === user.id
								? '#dcf8c6'
								: '#f1f0f0',
						maxWidth: '75%',
						wordWrap: 'break-word',
						padding: '10px',
						borderRadius: '18px',
						alignSelf:
							message.sender_id === user.id
								? 'flex-end'
								: 'flex-start',
						textAlign: 'left',
					}"
				>
					<div
						v-if="message.is_reply"
						class="reply-reference bg-light p-1 rounded mb-1"
					>
						<div class="text-muted">
							<div>
								<strong>
									{{ message.original_message.sender }}
								</strong>
							</div>
							<span class="original-message-content">
								{{
									message.original_message.content ||
									"This message was deleted"
								}}
							</span>
						</div>
					</div>
					<p v-if="message.is_forwarded" class="text-muted mb-0">
						<i class="bi bi-forward-fill px-1"></i><i>Forwarded</i>
					</p>
					<p
						v-if="
							conversation.conversation_type === 'group' &&
							message.sender_id !== user.id
						"
						style="
							font-weight: bold;
							margin-bottom: 2px;
							font-size: 16px;
						"
					>
						{{ message.sender_username }}
					</p>
					<p v-if="message.is_deleted" class="text-muted">
						<i>This message was deleted</i>
					</p>
					<div
						v-else
						style="
							display: flex;
							flex-direction: row;
							position: relative;
							gap: 10px;
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
						<div
							class="d-flex align-items-center gap-1 align-self-end"
						>
							<small
								class="text-muted"
								style="
									font-size: 12px;
									color: gray;
									margin-top: 4px;
									white-space: nowrap;
								"
							>
								{{ formatTimestamp(message.timestamp) }}
							</small>
							<i
								v-if="
									message.status === 'sent' &&
									message.sender_id == user.id
								"
								class="bi bi-check align-self-end"
								title="Sent"
							></i>
							<i
								v-if="
									message.status === 'read' &&
									message.sender_id == user.id
								"
								class="bi bi-check-all text-primary align-self-end"
								title="Read"
							></i>
						</div>
					</div>

					<!-- Reactions under messages -->
					<div
						v-if="message.reactions && message.reactions.length"
						:style="{
							position: 'absolute',
							bottom: '-20px',
							left:
								message.sender_id !== user.id ? '8px' : 'auto',
							right:
								message.sender_id === user.id ? '8px' : 'auto',
							background: 'white',
							padding: '3px 6px',
							borderRadius: '15px',
							boxShadow: '0px 2px 5px rgba(0, 0, 0, 0.1)',
							fontSize: '14px',
							cursor: 'pointer',
							zIndex: 100,
							display: 'flex',
							alignItems: 'center',
							gap: '10px',
						}"
						tabindex="0"
						role="button"
						data-bs-toggle="popover"
						@click="showReactionDetails($event, message)"
					>
						<span
							v-for="(count, emoji) in groupReactions(
								message.reactions
							)"
							:key="emoji"
							:style="{
								display: 'inline-flex',
								alignItems: 'center',
								justifyContent: 'center',
								width: '20px',
								height: '20px',
								fontSize: '16px',
								borderRadius: '50%',
								padding: '4px',
							}"
						>
							{{ emoji }}
							<span
								v-if="count > 1"
								:style="{
									fontSize: '12px',
									marginLeft: '6px',
									color: '#555',
								}"
								>{{ count }}</span
							>
						</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Message Input -->
		<div class="border-top">
			<div
				v-if="replyToMessage"
				class="d-flex align-items-center gap-2 p-2 pb-0"
			>
				<div class="text-muted p-1 m-0 rounded flex-grow-1">
					<div>
						<strong
							v-if="
								replyToMessage.sender_username === user.username
							"
						>
							Answering yourself
						</strong>
						<strong v-else>{{
							replyToMessage.sender_username
						}}</strong>
					</div>
					<div>
						<span v-if="replyToMessage.content">
							{{
								replyToMessage.content.length > 50
									? replyToMessage.content.substring(0, 50) +
									  "..."
									: replyToMessage.content
							}}
						</span>
						<span v-else-if="replyToMessage.photo_data"
							>Photo message</span
						>
					</div>
				</div>
				<div
					class="hover-bg rounded-circle d-flex align-items-center justify-content-center"
					style="
						width: 30px;
						height: 30px;
						overflow: hidden;
						font-size: 16px;
						line-height: 1;
					"
				>
					<i @click="cancelReply" class="bi bi-x"></i>
				</div>
			</div>

			<div class="d-flex align-items-end p-2">
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
				<button class="btn btn-primary ms-2" @click="sendMessage" title="Send Message">
					<img :src="SendIcon" alt="Send Message" width="24" />
				</button>
			</div>
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

	<!-- Forward Message Modal -->
	<div class="modal fade" id="forwardMessageModal" tabindex="-1">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">Forward Message</h5>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
					></button>
				</div>
				<div class="modal-body">
					<div
						v-if="availableConversations.length === 0"
						class="d-flex flex-column align-items-center p-4 text-muted"
					>
						<i class="bi bi-chat-dots fs-1 mb-2"></i>
						<p class="m-0">
							No conversation available to forward to.
						</p>
					</div>
					<div v-else class="p-3">
						<p class="mb-3 fw-bold">
							Select the conversation where you want to forward
							this message:
						</p>
						<ul class="list-group">
							<li
								v-for="conv in availableConversations"
								:key="conv.conversation_id"
								class="list-group-item d-flex align-items-center justify-content-between hover-bg"
								style="cursor: pointer"
								@click="forwardMessageTo(conv.conversation_id)"
							>
								<span>{{ conv.display_name }}</span>
								<i class="bi bi-arrow-right-circle"></i>
							</li>
						</ul>
					</div>
				</div>
				<div class="modal-footer">
					<button
						type="button"
						class="btn btn-secondary"
						data-bs-dismiss="modal"
					>
						Cancel
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Delete Message Confirmation Modal -->
	<div
		class="modal fade"
		id="deleteMessageModal"
		tabindex="-1"
		aria-labelledby="deleteMessageModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="deleteMessageModalLabel">
						Delete Message
					</h5>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
						aria-label="Close"
					></button>
				</div>
				<div class="modal-body">
					<p>Are you sure you want to delete this message?</p>
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
						@click="deleteMessage"
					>
						Delete Message
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.hover-bg:hover {
	background-color: var(--bs-gray-200);
	transition: background-color 0.3s ease;
	cursor: pointer;
}
.reply-reference {
	font-size: 0.9em;
	border-left: 3px solid #ddd;
}
</style>
