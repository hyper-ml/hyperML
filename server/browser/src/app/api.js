import RESTClient from './rest';
import storage from 'local-storage-fallback'


const routes = {
    serverInfo: () => '/info',
    login: () => '/login',
    getUserNotebooks: () => '/user/notebooks',
    createNotebook: () => '/user/notebooks/new',
    stopNotebook: (id) => `/user/notebooks/${id}/stop`,
    getResourceProfiles: () => '/resources/profiles'
}

class API {
    constructor(endpoint) {
        const namespace = 'Web'
        this.RESTClient = new RESTClient({
            endpoint, namespace
        })
    }

    makeCall(method, path, params){
        storage.setItem('token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7Ik5hbWUiOiJtaW5kaGFzaCJ9LCJpc3MiOiJoeXBlcmZsb3cifQ.EHXonVm1diWWEaL8oV8iHBHYLghukdl4sBsBL97Qnm8')
    

        return this.RESTClient.call(method, path, params, storage.getItem('token'))
        .catch(err => {
            if (err.status === 401) {
                storage.removeItem('token')
                throw new Error('Please re-login')
            }
            if (err.status) {
                throw new Error(`server returned an error [${err.status}]`)
            }
            throw new Error('HyperML server is unreachable')

        })
        .then(res => { 
            if (res.text) {
                let result = JSON.parse(res.text)
                let error =  result.error
                //let reason = result.reason
                if (error) {
                    //throw new Error(reason)
                    console.log('API error:', error)
                }
                return result
            } 
            console.log('ressponse:', res)
            if (res.statusCode > 201) {
                throw new Error("Failed to perform operation")
            } 
            
            return res.text
            
        })
    }

    LoggedIn() {
        return !!storage.getItem('token')
    }

    Login(args) {
        return this.makeCall('POST', routes.login(), args)
          .then(res => {
            storage.setItem('token', `${res.jwt}`)
            return res
          })
    }
    Logout() {
        storage.removeItem('token')
    }
    
    GetToken() {
        return storage.getItem('token')
    }

    ServerInfo() {
        return this.makeCall('GET', routes.serverInfo());
    }

    ListResourceProfiles() {
        return this.makeCall('GET', routes.getResourceProfiles())
   }

    ListNotebooks() {
         return this.makeCall('GET', routes.getUserNotebooks())
    }
    
    CreateNotebook(nb) {
        return this.makeCall('POST', routes.createNotebook(), nb)
    }

    StopNotebook(id) {
        return this.makeCall('PUT', routes.stopNotebook(id))
    }
}

const api = new API("http://localhost:8888/");

export default api; 