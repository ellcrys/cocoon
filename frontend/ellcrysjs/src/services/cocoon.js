// @flow
import Service from './service'
import uuid from 'uuid4'
import {FetchHandleResponse} from '../util/util'
import base64 from 'base-64'
import type {View} from '../flow_types/view'
import _ from 'lodash'
import {HTTPError, InvokeError} from '../errors/errors'

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
            })).then(FetchHandleResponse).then(function (data) {
                resolve(data)
            }).catch(function(err: HTTPError){
                reject(new InvokeError(err.status, err.body))
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
                console.error("failed to get manifest")
                reject(err)
            })
        })
    }
    
    /*
     * Get a view
     * 
     * @param {string} f                The function to invoke
     * @param {Array<string>} params    The function parameters
     * @returns {Promise<View|InvokeError>}
     * @memberof Cocoon
     */
    getView(f: string, params: Array<string>|void): Promise<View|InvokeError> {
        return new Promise((resolve, reject) => {
            this.invoke(uuid(), f, params).then(function (data) {
                let view = {}
                let err = _.attempt(() => {view = JSON.parse(base64.decode(data.body))})
                if (_.isError(err)) return reject(new Error("invalid view received"))
                return resolve(view)
            }).catch(function (err: InvokeError) {
                reject(err)
            })
        })
    }
}

export default Cocoon