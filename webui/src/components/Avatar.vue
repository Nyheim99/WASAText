<script>
import { ref, computed } from "vue";
import AvatarIcon from "../assets/person-circle.svg";

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

		const validateImage = (file) => {
      const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
      if (!allowedTypes.includes(file.type)) {
        validationMessage.value = "Only JPEG and PNG files are allowed.";
        return false;
      }
      return true;
    };

		const handleUpdatePhoto = async () => {
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

      uploading.value = true;
      const response = await props.updatePhoto(file);
      uploading.value = false;

      if (response.success) {
				cacheBuster.value = Date.now();
        const modal = document.getElementById("photoModal");
        const bootstrapModal = bootstrap.Modal.getInstance(modal);
        bootstrapModal.hide();
      } else {
        validationMessage.value = response.error;
        showValidation.value = true;
      }
    };

		const getAvatarSrc = computed(() => {
      if (!props.user.photoUrl || props.user.photoUrl === "") {
        return AvatarIcon;
      }
      return `http://localhost:3000${props.user.photoUrl}?t=${cacheBuster.value}`;
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
			handleUpdatePhoto
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
				:src="getAvatarSrc"
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
          data-bs-target="#photoModal"
        >
          Update Profile Picture
        </button>
      </li>
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
    id="photoModal"
    tabindex="-1"
    aria-labelledby="photoModalLabel"
    aria-hidden="true"
  >
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h1 class="modal-title fs-5" id="photoModalLabel">
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
              type="file"
              ref="fileInput"
              accept="image/jpeg, image/png"
              class="form-control"
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
            @click="handleUpdatePhoto"
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
