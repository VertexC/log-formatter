<template>
  <div id="agents">
    <div id="agent-info">
      <h1>Agents Monitor</h1>
      <table id="agent-table">
        <thead>
          <tr>
            <th>Id</th>
            <th>Status</th>
            <th>Address</th>
            <th>Configuration</th>
            <th>Refresh</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(agent,id) in agents" :key='id'>
            <td>{{ agent.id }}</td>
            <td v-if="agent.status == 'Unknown'">{{ agent.status }}(last alive {{ agent.alive }}s)</td>
            <td v-else>{{ agent.status }}</td>
            <td>{{ agent.address }}</td>
            <td><button v-bind:style="{'background-image': 'url(' + configIcon + ')'}" v-on:click="goConfig(id)"></button></td>
            <!-- <td><button class="config" v-on:click="showConfig(id)"></button></td> -->
            <td><button v-bind:style="{'background-image': 'url(' + refreshIcon + ')'}" v-on:click="refreshAgent(id)"></button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div>
      <div v-if="curId != null"> 
        <h1 > Configuration of Agent id: {{ curId }} </h1>
        <div><textarea v-model="config" placeholder="No Config Available" rows="10" cols="50"></textarea></div>
        <div><button v-on:click="updateConfig(curId, config)"></button></div>
      </div>
    </div>
    <div class="debug">
      <span>Response message:</span>
      <div> {{ message }} </div>
    </div>
    <div class="footer">
      <div>Icons made by <a href="https://www.flaticon.com/authors/freepik" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>
    </div>
    <router-view></router-view>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Agents',
  data() {
      return {
          message: null,
          config: null,
          curId: null,
          agents: {},
      }
  },
  computed: {
      refreshIcon: () => {
        return require('@/assets/reload.svg')
      },
      configIcon: () => {
        return require('@/assets/settings.svg')
      }
  },
  beforeRouteUpdate(to, from, next){
      this.refresh()
      next()
  },
  mounted() {
      this.refresh()
  },
  methods: {
    goConfig: function(id) {
      this.$router.push({name:"config", params: {id: id, config: this.agents[id].config}})
    },
    showConfig: function(id) {
      console.log(this.agents[id].config)
      this.config = this.agents[id].config
      this.curId = id
    },
    clearAgents: function() {
      this.agents = {}
    },
    refreshAgent: function(id) {
      console.log("Try to refresh Agent status:" + id.toString())
      return axios({
        method: 'get',
        url: this.$root.$serverUrl + '/agent',
        params: {
          id: id
        }
      }).then(response => {
        this.message = response
        var agent = JSON.parse(response['data']['agent'])
        this.agents[id] = agent
      }).catch((error) => {
        if (error.response.status == 503) {
          this.agents[id].status = "Unknown"
          this.message = error.response.data
        }
        alert(error)
      })
    },
    refresh: function() {
      console.log("refresh")
      return axios({
        method: 'get',
        url: this.$root.$serverUrl + '/app',
      }).then(response => {
        console.log(response)
        this.clearAgents()
        var agents = JSON.parse(response['data']['agents'])
        this.agents = agents
      }).catch((error) => {
        console.log(error)
        alert(error)
      })
    },
  },
}
</script>

<style>
:root {
  --color-form-highlight: #EEEEEE;
  --base-spacing-unit: 24px;
  --half-spacing-unit: var(--base-spacing-unit) / 2;
  --color-alpha: #1772FF;
  --color-form-highlight: #EEEEEE;
}

#agents {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
  padding:--var(base-spacing-unit);
	font-family:'Source Sans Pro', sans-serif;
	margin:0;
}

#agent-table {
    height: 80%;
    align-self: auto;
}

table {
    display: flex;
    flex-flow: column;
    height: 100%;
    width: 100%;
    border:1px solid var(--color-form-highlight);
}

table thead {
    /* head takes the height it requires, 
    and it's not scaled when table is resized */
    flex: 0 0 auto;
    width: calc(100% - 0.9em);
    background:#000;
		padding:(var(--half-spacing-unit) * 1.5) 0;
    color: var(--color-form-highlight);
}

table tbody {
    /* body takes all the remaining available space */
    flex: 1 1 auto;
    display: block;
    overflow-y: scroll;
}

table tbody tr {
    width: 100%;
    padding:(var(--half-spacing-unit) * 1.5) 0;
}

tbody tr:nth-of-type(odd) {
  background:var(--color-form-highlight);
}

table thead, table tbody tr {
  display: table;
  table-layout: fixed;
}

td {
  position:relative
}

tr button {
  position:absolute;
  max-width:100%;
  max-height:100%;
  top:0;
  height: 20px;
  width: 20px;
  padding: 0;
  margin: 0;
  text-align: center;
  border: none;
  background: transparent;
}

.footer {
  position: fixed;
  left: 0;
  bottom: 0;
  width: 100%;
  align-self: auto;
}
</style>
