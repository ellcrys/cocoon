// @flow
import Service from './service'
import uuid from 'uuid4'
import {FetchHandleResponse,FetchHandleJSON} from '../util/util'
import base64 from 'base-64'

// Cocoon defines a structure for interacting with a cocoon
class Cocoon extends Service {
    
    /**
     * Returns the address to the cocoon
     * TODO: retrieve from window.location.href
     * @returns URL to the cocoon
     * @memberof Cocoon
     */
    getAddress():string {
        return "http://localhost:8900"
    }
    
    /**
     * Sends an invoke request to the cocoon.
     * 
     * @param {String} id       A unique id for this request. UUID4 Expected
     * @param {String} func     The function to invoke
     * @param {Array}  params   A list of parameters to send to the cocoon code
     * @returns Promise
     * @memberof Cocoon
     */
    invoke(id: string, func: string, params: Array<string>|void) {
        return new Promise((resolve, reject) => {
            fetch(new Request(this.getAddress() + "/v1/invoke", { 
                method: 'POST', 
                headers: new Headers({ "Content-Type": "application/json" }),
                body: JSON.stringify({ id: id, "function": func, params: params })
            })).then(FetchHandleResponse).then(FetchHandleJSON).then(function (data) {
                resolve(data)
            }).catch(function(err){
                reject(err)
            })
        })
    }

    /**
     * Fetches the manifest file of the cocoon
     * @memberof Cocoon
     */
    getManifest(): Promise<{body: string}>{
        return new Promise((resolve, reject) => {
            this.invoke(uuid(), '@@GET_MANIFEST', undefined).then(function (data) {
                var manifest = JSON.parse(base64.decode(data.body))
                resolve(manifest)
            }).catch(function (err) {
                reject(new Error(`failed to get manifest: ${err.message}`))
            })
        })
    }
}

export default Cocoon