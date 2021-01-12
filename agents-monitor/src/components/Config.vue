<template>
  <div id="config">
    <div>
      <h1 > Configuration of Agent id: {{ agentId }} </h1>
      <div class="editor-container">
        <yaml-editor v-model="agentConfig" />
      </div>
      <div><button v-on:click="updateConfig(agentId, agentConfig)">Update Config</button></div>
    </div>
  </div>
</template>
    
<script>
import axios from 'axios'
import YamlEditor from '@/components/YamlEditor.vue';

export default {
  name: "Config",
  props: ['id', 'config'],
  components: { YamlEditor },
  data() {
      return {
          agentId: this.id,
          agentConfig: this.config,
      }
  },
  methods: {
    goBack() {
      window.history.length > 1 ? this.$router.go(-1) : this.$router.push('/')
    },
    updateConfig: function(id, config) {
      console.log("Try to update config")
      return axios({
        method: 'put',
        url: this.$root.$serverUrl + '/config',
        params: {
          id: id
        },
        data: {
          config: config
        }
      }).then(response => {
        console.log(response['data'])
        alert("Change config Succeed!")
      }).catch((error) => {
        alert(error)
      })
    },
  },
};
</script>

<style scoped>
textarea {
  max-height: 80%;
}

.editor-container{
  position: relative;
  height: 100%;
}
</style>