// @flow
import type {Manifest} from '../flow_types/manifest'

/**
 * Updates the cocoon manifest
 * @param {*} newManifest 
 */
export const updateManifest = (newManifest: Manifest) => {
    return {
        type: 'MANIFEST.UPDATE',
        payload: newManifest
    }
}