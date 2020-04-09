import SuperAgent from 'superagent-es6-promise';
import url from 'url';

export default class RESTClient {
    constructor(params) {
        this.endpoint = params.endpoint
        this.namespace = params.namespace
        this.version = '1.0'
        const parsedUrl = url.parse(this.endpoint)
        console.log('parsedURL:', parsedUrl)
        this.host = parsedUrl.hostname
        this.path = parsedUrl.path
        this.port = parsedUrl.port

        switch (parsedUrl.protocol) {
            case 'http:': {
                this.scheme = 'http'
                if (parsedUrl.port === 0) {
                    this.port = 80
                }
                break
            }
            case 'https:': {
                this.scheme = 'https'
                if (parsedUrl.port === 0) {
                  this.port = 443
                }
                break
            }
            default: {
                throw new Error('Unknown protocol: ' + parsedUrl.protocol)
            }
        }
    }

    getEndpoint() {
        if (this.endpoint) {
            return this.endpoint
        } else {
            return this.scheme + "://" + this.host + ":" + this.port
        }
    }

    call(method, path, params, token) {
        
        if (path.charAt(0)=== '/') {
            path = path.slice(1, path.length)
        }

        if (!method) {
            method = 'GET'
        }
        if (!params) {
            params = {}
        }

        let reqParams = {
            host: this.host,
            port: this.port,
            path: path,
            method: method,
            scheme: this.scheme,
            headers: {
                'Content-Type': 'application/json'
            }
        }

        if (token) {
            reqParams.headers.Authorization = 'Bearer ' + token
        }

        let req; 

        if (method === 'GET') {
            req = SuperAgent.get(this.getEndpoint() + path)
        } else if (method === 'POST') {
            req = SuperAgent.post(this.getEndpoint() + path)
        } else if (method === 'PUT') {
            req = SuperAgent.put(this.getEndpoint() + path)
        } else if (method === 'DELETE') {
            req = SuperAgent.delete(this.getEndpoint() + path)
        } 

        for (let key in reqParams.headers) {
            req.set(key, reqParams.headers[key])
        }
  
        return req.send(JSON.stringify(params))
        
      
    }
}