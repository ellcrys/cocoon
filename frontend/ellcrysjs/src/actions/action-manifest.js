// @flow

/**
 * Updates the cocoon manifest
 * @param {*} newManifest 
 */
export const updateManifest = (newManifest: {}) => {
    return {
        type: 'MANIFEST.UPDATE',
        payload: newManifest
    }
}