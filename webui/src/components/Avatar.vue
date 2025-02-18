<script>
import { ref, computed } from "vue";
import AvatarIcon from "/person-fill.svg";

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
	setup(props) {
		const cacheBuster = ref(Date.now());
		const newUsername = ref("");
		const fileInput = ref(null);
		const uploading = ref(false);
		const showValidation = ref(false);
		const validationMessage = ref("");

		//Validate new username
		const validateUsername = (username) => {
			if (username.length < 3 || username.length > 16) {
				validationMessage.value =
					"Username must be between 3 and 16 characters long.";
				return false;
			}
			if (!/^[a-zA-Z0-9]*$/.test(username)) {
				validationMessage.value =
					"Username can only contain alphanumeric characters.";
				return false;
			}
			return true;
		};

		//Updating the user's username
		const handleUpdateUsername = async () => {
			//Validate the input
			if (!validateUsername(newUsername.value)) {
				showValidation.value = true;
				return;
			}

			//Update the username in HomeView
			const response = await props.updateUsername(newUsername.value);

			//Display success or failure depending on response
			if (response === true) {
				newUsername.value = "";
				showValidation.value = false;
				const modal = document.getElementById("usernameModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();
			} else {
				validationMessage.value = response;
				showValidation.value = true;
			}
		};

		//Validate the new profile picture
		const validateImage = (file) => {
			const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
			if (!allowedTypes.includes(file.type)) {
				validationMessage.value =
					"Only JPEG and PNG files are allowed.";
				return false;
			}
			return true;
		};

		//Update the user's profile picture
		const handleUpdatePhoto = async () => {
			//Validate the input
			const file = fileInput.value?.files[0];
			if (!file) {
				validationMessage.value = "Please select a file.";
				showValidation.value = true;
				return;
			}

			if (!validateImage(file)) {
				showValidation.value = true;
				return;
			}

			//Upload the file in HomeView
			uploading.value = true;
			const response = await props.updatePhoto(file);
			uploading.value = false;

			//Display success or failure depending on response
			if (response === true) {
				cacheBuster.value = Date.now();
				const modal = document.getElementById("photoModal");
				const bootstrapModal = bootstrap.Modal.getInstance(modal);
				bootstrapModal.hide();
			} else {
				validationMessage.value = response.error;
				showValidation.value = true;
			}
		};

		//Return the display photo for the avatar
		const getAvatarSrc = computed(() => {
			if (!props.user.photo_url || props.user.photo_url === "") {
				return AvatarIcon;
			}
			return `${__API_URL__}${props.user.photo_url}?t=${cacheBuster.value}`;
		});

		return {
			AvatarIcon,
			getAvatarSrc,
			newUsername,
			fileInput,
			uploading,
			showValidation,
			validationMessage,
			handleUpdateUsername,
			handleUpdatePhoto,
		};
	},
};
</script>

<template>
	<!-- Avatar -->
	<div class="dropup">
		<button
			class="btn btn-light shadow-sm p-0 align-items-center justify-content-center"
			type="button"
			data-bs-toggle="dropdown"
			data-bs-auto-close="outside"
			data-bs-offset="0, 10"
			title="User Actions"
			aria-expanded="false"
			style="
				width: 60px;
				height: 60px;
				border-radius: 50%;
				overflow: hidden;
			"
		>
			<img
				:src="getAvatarSrc"
				alt="User Avatar"
				style="width: 100%; height: 100%; object-fit: cover"
			/>
		</button>

		<!-- Dropup menu -->
		<ul class="dropdown-menu">
			<!-- Display logged in user's username -->
			<li>
				<span class="dropdown-item-text"> @{{ user.username }} </span>
			</li>
			<li><hr class="dropdown-divider" /></li>

			<!-- Uploading/Changing profile picture -->
			<li>
				<button
					class="dropdown-item"
					type="button"
					data-bs-toggle="modal"
					data-bs-target="#photoModal"
				>
					Update Profile Picture
				</button>
			</li>

			<!-- Uploading/Changing username -->
			<li>
				<button
					class="dropdown-item"
					type="button"
					data-bs-toggle="modal"
					data-bs-target="#usernameModal"
				>
					Update Username
				</button>
			</li>
			<li><hr class="dropdown-divider" /></li>

			<!-- Logout button -->
			<li>
				<button class="dropdown-item" type="button" @click="logout">
					Log out
				</button>
			</li>
		</ul>
	</div>

	<!-- Modal to upload new profile picture -->
	<div
		id="photoModal"
		class="modal fade"
		tabindex="-1"
		aria-labelledby="photoModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h1 id="photoModalLabel" class="modal-title fs-5">
						Upload a new profile picture
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
							ref="fileInput"
							type="file"
							accept="image/jpeg, image/png"
							class="form-control"
						/>
					</div>
					<p
						v-if="showValidation"
						class="text-danger small mt-2"
						aria-live="assertive"
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
						:disabled="uploading"
						@click="handleUpdatePhoto"
					>
						{{ uploading ? "Uploading..." : "Upload" }}
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Modal to update the user's username -->
	<div
		id="usernameModal"
		class="modal fade"
		tabindex="-1"
		aria-labelledby="usernameModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h1 id="usernameModalLabel" class="modal-title fs-5">
						Insert your new username
					</h1>
					<button
						type="button"
						class="btn-close"
						data-bs-dismiss="modal"
						aria-label="Close"
					></button>
				</div>
				<div class="modal-body">
					<p class="text-muted small mb-3">
						Username must be 3–16 characters long and contain only
						alphanumeric characters (letters and numbers).
					</p>
					<div class="input-group flex-nowrap">
						<span id="addon-wrapping" class="input-group-text"
							>@</span
						>
						<input
							v-model="newUsername"
							type="text"
							class="form-control"
							placeholder="Username"
							aria-label="username"
							aria-describedby="addon-wrapping"
						/>
					</div>
					<p
						v-if="showValidation"
						class="text-danger small mt-2"
						aria-live="assertive"
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
						@click="handleUpdateUsername"
					>
						Update username
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
