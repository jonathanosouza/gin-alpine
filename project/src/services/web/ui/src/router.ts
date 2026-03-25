import { createRouter, createWebHistory } from "vue-router";
import { getAppContext } from "./lib/appContext";
import ForgotPasswordPage from "./pages/ForgotPasswordPage.vue";
import LoginPage from "./pages/LoginPage.vue";
import ProfilePage from "./pages/ProfilePage.vue";
import ResetPasswordPage from "./pages/ResetPasswordPage.vue";
import UsersPage from "./pages/UsersPage.vue";
import ComercialPage from "./pages/ComercialPage.vue";
import LogisticaPage from "./pages/LogisticaPage.vue";
import FinanceiroPage from "./pages/FinanceiroPage.vue";
import ConfigHome from "./pages/ConfigHome.vue";
import ClientePage from "./pages/ClientePage.vue";
import VendedorPage from "./pages/VendedorPage.vue";
import RealMetaPage from "./pages/RealMetaPage.vue";
import FornecedorPage from "./pages/FornecedorPage.vue";
import VendasPeriodoPage from "./pages/VendasPeriodoPage.vue";
import CrescimentoAnoPage from "./pages/CrescimentoAnoPage.vue";
import EvolucaoVendedorPage from "./pages/EvolucaoVendedorPage.vue";
import EstoquePage from "./pages/EstoquePage.vue";
import SugestaoComprasPage from "./pages/SugestaoComprasPage.vue";
import CurvaABCPage from "./pages/CurvaABCPage.vue";
import ContasPagarPage from "./pages/ContasPagarPage.vue";
import ContasReceberPage from "./pages/ContasReceberPage.vue";
import FluxoCaixaPage from "./pages/FluxoCaixaPage.vue";
import DREPage from "./pages/DREPage.vue";
import MetasPage from "./pages/MetasPage.vue";
import AutomacaoPage from "./pages/AutomacaoPage.vue";
import VisaoGeralPage from "./pages/VisaoGeralPage.vue";
import VisaoGeralCalculosPage from "./pages/VisaoGeralCalculosPage.vue";

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
    { path: "/", redirect: "/visao-geral" },
    { path: "/visao-geral", component: VisaoGeralPage },
    { path: "/visao-geral/calculos", component: VisaoGeralCalculosPage },
    { path: "/comercial", component: ComercialPage },
    { path: "/comercial/cliente", component: ClientePage },
    { path: "/comercial/vendedor", component: VendedorPage },
    { path: "/cliente", redirect: "/comercial/cliente" },
    { path: "/vendedor", redirect: "/comercial/vendedor" },
    { path: "/comercial/real-meta", component: RealMetaPage },
    { path: "/comercial/fornecedor", component: FornecedorPage },
    { path: "/comercial/vendas-periodo", component: VendasPeriodoPage },
    { path: "/comercial/crescimento-ano", component: CrescimentoAnoPage },
    { path: "/comercial/evolucao-vendedor", component: EvolucaoVendedorPage },
    { path: "/logistica", component: LogisticaPage },
    { path: "/logistica/estoque", component: EstoquePage },
    { path: "/logistica/sugestao-compras", component: SugestaoComprasPage },
    { path: "/logistica/curva-abc", component: CurvaABCPage },
    { path: "/financeiro", component: FinanceiroPage },
    { path: "/financeiro/contas-pagar", component: ContasPagarPage },
    { path: "/financeiro/contas-receber", component: ContasReceberPage },
    { path: "/financeiro/fluxo-caixa", component: FluxoCaixaPage },
    { path: "/financeiro/dre", component: DREPage },
    { path: "/configuracao", component: ConfigHome },
    { path: "/configuracao/usuarios", component: UsersPage },
    { path: "/configuracao/metas", component: MetasPage },
    { path: "/configuracao/automacao", component: AutomacaoPage },
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
