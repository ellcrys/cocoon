// @flow
export default function (state: {} | null = null, action: { type: string, payload: {} }) {
    switch (action.type) {
        case "MANIFEST.UPDATE":
            return action.payload
        default:
            return state
    }
}