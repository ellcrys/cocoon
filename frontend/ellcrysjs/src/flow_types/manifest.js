// @flow

export type Func = {
    title: string,
    name: string
}

export type Manifest = {
    name: string,
    description: string,
    functions: Array<Func>
}