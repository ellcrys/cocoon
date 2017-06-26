import {bindActionCreators} from 'redux';

export function mapStateToProps(state) {
    return {
        manifest: state.manifest
    };
}

export function matchDispatchToProps(dispatch){
    return bindActionCreators({}, dispatch);
}

