// @flow
import {HTTPError} from '../errors/errors'

export function AddToClassName(existing: string, newClass: string): string {
    existing = existing || ''
    let existingList = existing.split(" ")
    existingList.push(newClass)
    return existingList.join(" ")
}

export function AddToInlineStyle(existing: string, newStyle: string): string {
    existing = existing || ''
    let existingList = existing.split(";")
    existingList.push(newStyle)
    return existingList.join(";")
}

// FetchHandleResponse handles response from fetch. It returns 
// a resolved promise if response code is between 200-299 or a rejected
// rejected promise if otherwise.
export async function FetchHandleResponse(response: Response): Promise<any> { 
    let resp = await response.json()
    if (response.ok) {
        return Promise.resolve(resp)
    }
    return Promise.reject(new HTTPError(response.status, response.statusText, resp))
}