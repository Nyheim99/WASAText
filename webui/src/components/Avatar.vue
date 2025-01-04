<script>
import { ref } from "vue";

export default {
	props: {
		src: {
			type: String,
			required: true,
		},
		logout: Function,
		updateUsername: Function,
		user: {
			type: Object,
			required: true,
		},
	},
	setup(props) {
		const newUsername = ref("");
		const showValidation = ref(false);
		const validationMessage = ref("");

		const validateUsername = (username) => {
			if (username.trim() === "") {
				validationMessage.value = "Username cannot be empty.";
				return false;
			}
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

		const handleUpdateUsername = async () => {
			if (!validateUsername(newUsername.value)) {
				showValidation.value = true;
				return;
			}

			const response = await props.updateUsername(newUsername.value);

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

		return {
			newUsername,
			showValidation,
			validationMessage,
			handleUpdateUsername,
		};
	},
};
</script>

<template>
	<div class="dropup">
		<button
			class="btn btn-light shadow-sm p-0 align-items-center justify-content-center"
			type="button"
			data-bs-toggle="dropdown"
			data-bs-auto-close="outside"
			data-bs-offset="0, 10"
			aria-expanded="false"
			style="
				width: 60px;
				height: 60px;
				border-radius: 50%;
				overflow: hidden;
			"
		>
			<img
				:src="src"
				alt="User Avatar"
				style="width: 100%; height: 100%; object-fit: cover"
			/>
		</button>
		<ul class="dropdown-menu">
			<li>
				<span class="dropdown-item-text"> @{{ user.username }} </span>
			</li>
			<li><hr class="dropdown-divider" /></li>
			<li>
				<button
					class="dropdown-item"
					type="button"
					data-bs-toggle="modal"
					data-bs-target="#usernameModal"
				>
					Update username
				</button>
			</li>
			<li><hr class="dropdown-divider" /></li>
			<li>
				<button class="dropdown-item" type="button" @click="logout">
					Log out
				</button>
			</li>
		</ul>
	</div>

	<div
		class="modal fade"
		id="usernameModal"
		tabindex="-1"
		aria-labelledby="usernameModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h1 class="modal-title fs-5" id="usernameModalLabel">
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
						<span class="input-group-text" id="addon-wrapping"
							>@</span
						>
						<input
							type="text"
							v-model="newUsername"
							class="form-control"
							placeholder="Username"
							aria-label="username"
							aria-describedby="addon-wrapping"
						/>
					</div>
					<p
						class="text-danger small mt-2"
						aria-live="assertive"
						v-if="showValidation"
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
