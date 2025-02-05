import axios from "axios";

const instance = axios.create({
	baseURL: `${__API_URL__}`,
	timeout: 1000 * 5,
	headers: {
		"Content-Type": "application/json",
	},
});

instance.interceptors.request.use((config) => {
	const userId = localStorage.getItem("userId");
	if (userId) {
		config.headers.Authorization = `Bearer ${userId}`;
	}
	return config;
});

export default instance;