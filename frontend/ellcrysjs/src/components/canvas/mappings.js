// @flow
import {bindActionCreators} from 'redux';
import {updateManifest} from '../../actions/action-manifest'

export function mapStateToProps(state: any) {
    return {
        manifest: state.manifest
    };
}

export function matchDispatchToProps(dispatch: any){
    return bindActionCreators({updateManifest: updateManifest}, dispatch);
}

