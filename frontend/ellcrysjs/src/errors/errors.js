// @flow

export class HTTPError extends Error {
    status: number
    statusText: string
    body: any
    constructor(status: number, statusText: string, body: any) {
        super(statusText)
        this.name = this.constructor.name;
        Error.captureStackTrace(this, HTTPError)
        this.status = status
        this.statusText = statusText
        this.body = body
    }
}

type InvokeErrorBody = {
    code: string, 
    error: boolean, 
    msg: string
}

export class InvokeError extends Error {
    status: number
    body: InvokeErrorBody
    constructor(status: number, body: InvokeErrorBody) {
        super(body.msg)
        this.name = this.constructor.name;
        Error.captureStackTrace(this, InvokeError)
        this.status = status
        this.body = body
    }
}
