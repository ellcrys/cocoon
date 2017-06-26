// @flow

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
export function FetchHandleResponse(response: Response): Promise<any> { 
    if (response.ok) {
        return Promise.resolve(response)
    }
    return Promise.reject(new Error(response.statusText))
}

// FetchHandleJSON parse the json body returned by fetch API
export function FetchHandleJSON(response: Response): Promise<any> {  
    return response.json()  
}