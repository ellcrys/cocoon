import {bindActionCreators} from 'redux';
import {updateManifest} from '../../actions/action-manifest'

export function mapStateToProps(state) {
    return {
        manifest: state.manifest
    };
}

export function matchDispatchToProps(dispatch){
    return bindActionCreators({updateManifest: updateManifest}, dispatch);
}

