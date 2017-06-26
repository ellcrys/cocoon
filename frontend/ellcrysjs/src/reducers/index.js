import {combineReducers} from 'redux'
import ManifestReducer from './reducer-manifest'

const all = combineReducers({
    manifest: ManifestReducer
})

export default all