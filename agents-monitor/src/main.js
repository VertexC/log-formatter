import Vue from 'vue'
import App from './App.vue'
import { TableComponent, TableColumn } from 'vue-table-component';
import Config from './components/Config.vue'
import Agents from './components/Agents.vue'
import VueRouter from 'vue-router'
Vue.use(VueRouter)

Vue.component('table-component', TableComponent);
Vue.component('table-column', TableColumn);
Vue.component('config', Config)

Vue.config.productionTip = false

const routes = [
    {
        name: "agents",
        path: "/agents",
        component: Agents,
    },
    {
        name: "config",
        path: "/config",
        component: Config,
        props: true,
    },
]

const router = new VueRouter({
    routes: routes    
})

router.replace('/agents')

const originalPush = VueRouter.prototype.push;
VueRouter.prototype.push = function push(location) {
    return originalPush.call(this, location).catch(err => err)
}

new Vue({
    router: router,
    render: h=> h(App),
}).$mount('#app')