import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from "../views/LoginView.vue"

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: HomeView},
		{path: "/login", component: LoginView}
	]
})

//Redirect user to login if not logged in
router.beforeEach((to, from, next) => {
	const isLoggedIn = !!localStorage.getItem("userId");

	//If user tries to acces /login while being logged in
	//redirect to /
	if (to.path === "/login" && isLoggedIn) {
		next("/");
	//If user is not logged in, prompt to do so.
	} else if (!isLoggedIn && to.path !== "/login") {
		next("/login");
	//Otherwise let user proceed as intended
	} else {
		next();
	}
});

export default router
