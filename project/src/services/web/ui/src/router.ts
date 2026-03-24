import { createRouter, createWebHistory } from "vue-router";
import { getAppContext } from "./lib/appContext";
import DashboardHome from "./pages/DashboardHome.vue";
import ForgotPasswordPage from "./pages/ForgotPasswordPage.vue";
import LoginPage from "./pages/LoginPage.vue";
import ProfilePage from "./pages/ProfilePage.vue";
import ResetPasswordPage from "./pages/ResetPasswordPage.vue";
import UsersPage from "./pages/UsersPage.vue";
import ComercialPage from "./pages/ComercialPage.vue";
import LogisticaPage from "./pages/LogisticaPage.vue";
import FinanceiroPage from "./pages/FinanceiroPage.vue";
import ConfigHome from "./pages/ConfigHome.vue";

const publicPaths = new Set<string>([
  "/login",
  "/recuperar-senha",
]);

export const router = createRouter({
  history: createWebHistory(
    typeof window !== "undefined" && window.location.pathname.startsWith("/static/vue/") ? "/static/vue/" : "/",
  ),
  routes: [
    { path: "/login", component: LoginPage },
    { path: "/recuperar-senha", component: ForgotPasswordPage },
    { path: "/reset-senha/:uuid", component: ResetPasswordPage },
    { path: "/", component: DashboardHome },
    { path: "/comercial", component: ComercialPage },
    { path: "/logistica", component: LogisticaPage },
    { path: "/financeiro", component: FinanceiroPage },
    { path: "/configuracao", component: ConfigHome },
    { path: "/configuracao/usuarios", component: UsersPage },
    { path: "/perfil", component: ProfilePage },
  ],
});

router.beforeEach((to) => {
  const ctx = getAppContext();
  if (publicPaths.has(to.path) || to.path.startsWith("/reset-senha/")) {
    return true;
  }
  if (ctx.IsAuth) {
    return true;
  }
  return { path: "/login" };
});
